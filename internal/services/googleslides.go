package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gocreator/internal/interfaces"

	"github.com/spf13/afero"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/slides/v1"
)

const (
	// defaultHTTPTimeout is the timeout for HTTP requests when downloading slide images
	defaultHTTPTimeout = 30 * time.Second
	
	// OAuth 2.0 token file for storing and refreshing access tokens
	tokenFile = ".gocreator-token.json"
)

// GoogleSlidesService handles fetching slides and notes from Google Slides
type GoogleSlidesService struct {
	fs     afero.Fs
	logger interfaces.Logger
}

// NewGoogleSlidesService creates a new Google Slides service
func NewGoogleSlidesService(fs afero.Fs, logger interfaces.Logger) *GoogleSlidesService {
	return &GoogleSlidesService{
		fs:     fs,
		logger: logger,
	}
}

// LoadFromGoogleSlides fetches slides as images and their speaker notes from a Google Slides presentation
func (s *GoogleSlidesService) LoadFromGoogleSlides(ctx context.Context, presentationID, outputDir string) ([]string, []string, error) {
	// Create Google Slides service
	slidesService, err := s.createSlidesService(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create slides service: %w", err)
	}

	// Get presentation
	presentation, err := slidesService.Presentations.Get(presentationID).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get presentation: %w", err)
	}

	s.logger.Info("Fetched presentation", "title", presentation.Title, "slideCount", len(presentation.Slides))

	// Create output directory
	if err := s.fs.MkdirAll(outputDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var slidePaths []string
	var notes []string

	// Process each slide
	for i, slide := range presentation.Slides {
		s.logger.Debug("Processing slide", "index", i, "objectId", slide.ObjectId)

		// Get slide thumbnail using the API
		thumbnail, err := slidesService.Presentations.Pages.GetThumbnail(presentationID, slide.ObjectId).Do()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get thumbnail for slide %d: %w", i, err)
		}

		// Download slide image from thumbnail URL
		slidePath := filepath.Join(outputDir, fmt.Sprintf("slide_%d.png", i+1))
		if err := s.downloadImage(ctx, thumbnail.ContentUrl, slidePath); err != nil {
			return nil, nil, fmt.Errorf("failed to download slide %d: %w", i+1, err)
		}

		slidePaths = append(slidePaths, slidePath)

		// Extract speaker notes
		note := s.extractSpeakerNotes(slide)
		notes = append(notes, note)

		s.logger.Debug("Processed slide", "index", i, "path", slidePath, "noteLength", len(note))
	}

	return slidePaths, notes, nil
}

// createSlidesService creates a Google Slides API service with credentials
// Supports both OAuth 2.0 (user authorization) and service account credentials
func (s *GoogleSlidesService) createSlidesService(ctx context.Context) (*slides.Service, error) {
	// Try OAuth 2.0 credentials first (for user authorization)
	oauthCredPath := os.Getenv("GOOGLE_OAUTH_CREDENTIALS")
	if oauthCredPath != "" {
		s.logger.Debug("Using OAuth 2.0 credentials", "path", oauthCredPath)
		return s.createSlidesServiceWithOAuth(ctx, oauthCredPath)
	}
	
	// Fall back to service account credentials
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath != "" {
		s.logger.Debug("Using service account credentials", "path", credentialsPath)
		service, err := slides.NewService(ctx, option.WithCredentialsFile(credentialsPath))
		if err != nil {
			return nil, fmt.Errorf("failed to create slides service with service account: %w", err)
		}
		return service, nil
	}
	
	return nil, fmt.Errorf("no Google credentials found. Set either GOOGLE_OAUTH_CREDENTIALS (for OAuth 2.0) or GOOGLE_APPLICATION_CREDENTIALS (for service account). See GOOGLE_SLIDES_GUIDE.md for setup instructions")
}

// createSlidesServiceWithOAuth creates a Google Slides API service using OAuth 2.0
func (s *GoogleSlidesService) createSlidesServiceWithOAuth(ctx context.Context, credentialsPath string) (*slides.Service, error) {
	// Read OAuth 2.0 credentials file (client ID and secret)
	credData, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OAuth credentials file: %w", err)
	}
	
	// Parse OAuth 2.0 config
	config, err := google.ConfigFromJSON(credData, 
		"https://www.googleapis.com/auth/presentations",
		"https://www.googleapis.com/auth/drive.file",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OAuth credentials: %w", err)
	}
	
	// Get or refresh access token
	token, err := s.getToken(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %w", err)
	}
	
	// Create HTTP client with token
	client := config.Client(ctx, token)
	
	// Create Google Slides service
	service, err := slides.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create slides service with OAuth: %w", err)
	}
	
	return service, nil
}

// getToken retrieves a token from file or initiates OAuth flow
func (s *GoogleSlidesService) getToken(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Try to load token from file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	tokenPath := filepath.Join(homeDir, ".config", "gocreator", tokenFile)
	token, err := s.loadToken(tokenPath)
	if err == nil {
		// Token loaded successfully, check if it's valid or refresh it
		if token.Valid() {
			s.logger.Debug("Using cached OAuth token")
			return token, nil
		}
		
		// Try to refresh the token
		s.logger.Debug("Refreshing OAuth token")
		tokenSource := config.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err == nil {
			// Save refreshed token
			if saveErr := s.saveToken(tokenPath, newToken); saveErr != nil {
				s.logger.Error("Failed to save refreshed token", "error", saveErr)
			}
			return newToken, nil
		}
		s.logger.Debug("Failed to refresh token, will request new authorization", "error", err)
	}
	
	// No valid token, initiate OAuth flow
	s.logger.Info("No valid OAuth token found. Initiating authorization flow...")
	token, err = s.getTokenFromWeb(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from web: %w", err)
	}
	
	// Save token for future use
	if err := s.saveToken(tokenPath, token); err != nil {
		s.logger.Error("Failed to save OAuth token", "error", err)
	}
	
	return token, nil
}

// getTokenFromWeb initiates the OAuth 2.0 authorization flow
func (s *GoogleSlidesService) getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Generate authorization URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	
	fmt.Printf("\n🔐 Google Slides Authorization Required\n")
	fmt.Printf("═══════════════════════════════════════\n\n")
	fmt.Printf("Please visit this URL to authorize this application:\n\n")
	fmt.Printf("%s\n\n", authURL)
	fmt.Printf("After authorization, you will receive an authorization code.\n")
	fmt.Printf("Enter the authorization code here: ")
	
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("failed to read authorization code: %w", err)
	}
	
	// Exchange authorization code for token
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code for token: %w", err)
	}
	
	fmt.Printf("\n✓ Authorization successful!\n\n")
	
	return token, nil
}

// loadToken loads an OAuth token from a file
func (s *GoogleSlidesService) loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	token := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, err
	}
	
	return token, nil
}

// saveToken saves an OAuth token to a file
func (s *GoogleSlidesService) saveToken(path string, token *oauth2.Token) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}
	
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()
	
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}
	
	return nil
}

// downloadImage downloads an image from a URL and saves it to the filesystem
func (s *GoogleSlidesService) downloadImage(ctx context.Context, url, outputPath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: defaultHTTPTimeout,
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Create output file
	file, err := s.fs.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy image data to file
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	return nil
}

// extractSpeakerNotes extracts speaker notes from a slide
func (s *GoogleSlidesService) extractSpeakerNotes(slide *slides.Page) string {
	if slide.SlideProperties == nil || slide.SlideProperties.NotesPage == nil {
		return ""
	}

	notesPage := slide.SlideProperties.NotesPage
	var noteText string

	// Iterate through page elements to find text in notes
	for _, pageElement := range notesPage.PageElements {
		if pageElement.Shape != nil && pageElement.Shape.Text != nil {
			for _, textElement := range pageElement.Shape.Text.TextElements {
				if textElement.TextRun != nil {
					noteText += textElement.TextRun.Content
				}
			}
		}
	}

	return noteText
}

// LoadSlides is not implemented for Google Slides service
func (s *GoogleSlidesService) LoadSlides(ctx context.Context, dir string) ([]string, error) {
	return nil, fmt.Errorf("LoadSlides not implemented for GoogleSlidesService, use LoadFromGoogleSlides instead")
}
