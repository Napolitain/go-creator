package text

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/openai/openai-go/v3"
)

// SlideText is the struct for the slide text content
type SlideText struct {
	Text string
}

// NewSlideText creates a new SlideText
func NewSlideText(text string) *SlideText {
	return &SlideText{Text: text}
}

// Hash returns the hash of the slide text
func (s *SlideText) Hash() string {
	hash := sha256.New()
	hash.Write([]byte(s.Text))
	return hex.EncodeToString(hash.Sum(nil))
}

// Text is an array of SlideText, corresponding basically to a full video.
type Text struct {
	Lang       string
	RootDir    string
	DataDir    string
	CacheDir   string
	LangDir    string
	AudioDir   string
	TextDir    string
	OpenAI     openai.Client
	SlidesText []*SlideText
	Hashes     []string
}

// NewText creates a new Text struct. It corresponds to an entire video text (single language).
func NewText(rootDir, lang string, client openai.Client, logger *slog.Logger) *Text {
	dataDir := filepath.Join(rootDir, "data")
	cacheDir := filepath.Join(dataDir, "cache")
	langDir := filepath.Join(cacheDir, lang)
	audioDir := filepath.Join(langDir, "audio")
	textDir := filepath.Join(langDir, "text")

	dirs := []string{cacheDir, langDir, audioDir, textDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.Mkdir(dir, os.ModePerm)
			if err != nil {
				logger.Error("Failed to create directory", "dir", dir, "error", err)
				os.Exit(1)
			}
		}
	}

	return &Text{
		Lang:       lang,
		RootDir:    rootDir,
		DataDir:    dataDir,
		CacheDir:   cacheDir,
		LangDir:    langDir,
		AudioDir:   audioDir,
		TextDir:    textDir,
		OpenAI:     client,
		SlidesText: []*SlideText{},
		Hashes:     []string{},
	}
}

// GenerateText loads the text from the text file.
func (t *Text) GenerateText(inputText *Text, logger *slog.Logger) error {
	var textFilePath string
	if inputText == nil {
		textFilePath = filepath.Join(t.DataDir, "texts.txt")
	} else {
		textFilePath = filepath.Join(t.TextDir, "texts.txt")
	}

	textFile, err := os.Open(textFilePath)
	if err != nil {
		return err
	}
	defer textFile.Close()

	if inputText != nil {
		t.translateText(inputText, logger)
	} else {
		t.loadTextInput(textFile)
	}

	return nil
}

func (t *Text) translateText(inputText *Text, logger *slog.Logger) {
	var wg sync.WaitGroup
	t.SlidesText = make([]*SlideText, len(inputText.SlidesText))
	t.Hashes = make([]string, len(inputText.SlidesText))
	for i, slideText := range inputText.SlidesText {
		wg.Add(1)
		go func(i int, slideText *SlideText) {
			defer wg.Done()

			resp, err := t.OpenAI.Chat.Completions.New(
				context.Background(),
				openai.ChatCompletionNewParams{
					Model: openai.ChatModelGPT4oMini,
					Messages: []openai.ChatCompletionMessageParamUnion{
						openai.UserMessage(fmt.Sprintf("Translate '%s' to %s and don't return anything else than the translation.", slideText.Text, t.Lang)),
					},
				},
			)
			if err != nil {
				logger.Error("Translation error", "error", err)
				os.Exit(1)
			}

			translatedText := resp.Choices[0].Message.Content
			t.SlidesText[i] = NewSlideText(translatedText)
			t.Hashes[i] = inputText.Hashes[i]
		}(i, slideText)
	}
	wg.Wait()
	t.saveTextFile(t.SlidesText, logger)
}

func (t *Text) loadTextInput(textFile *os.File) {
	scanner := bufio.NewScanner(textFile)
	slideText := ""
	t.Hashes = []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "-" {
			newSlideText := NewSlideText(slideText)
			t.SlidesText = append(t.SlidesText, newSlideText)
			t.Hashes = append(t.Hashes, newSlideText.Hash())
			slideText = ""
		} else {
			slideText += line + "\n"
		}
	}
	t.SlidesText = append(t.SlidesText, NewSlideText(slideText))
	t.Hashes = append(t.Hashes, t.SlidesText[len(t.SlidesText)-1].Hash())
}

func (t *Text) saveTextFile(slidesText []*SlideText, logger *slog.Logger) error {
	textFile := filepath.Join(t.TextDir, "texts.txt")
	hashFile := filepath.Join(t.TextDir, "hashes")

	textF, err := os.Create(textFile)
	if err != nil {
		return err
	}
	defer textF.Close()

	hashF, err := os.Create(hashFile)
	if err != nil {
		return err
	}
	defer hashF.Close()

	for i, slideText := range slidesText {
		if i == len(slidesText)-1 {
			textF.WriteString(slideText.Text)
			hashF.WriteString(t.Hashes[i])
		} else {
			textF.WriteString(slideText.Text + "\n-\n")
			hashF.WriteString(t.Hashes[i] + "\n")
		}
	}
	return nil
}

// GenerateCacheHashes generates hashes from cache
func (t *Text) GenerateCacheHashes(directory string) []string {
	hashFile := filepath.Join(directory, "hashes")
	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		return make([]string, len(t.SlidesText))
	}

	file, err := os.Open(hashFile)
	if err != nil {
		return make([]string, len(t.SlidesText))
	}
	defer file.Close()

	var hashes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hashes = append(hashes, scanner.Text())
	}
	return hashes
}

// Texts struct is a set of texts for different languages
type Texts struct {
	LangIn   string
	LangsOut []string
	DataDir  string
	Texts    []*Text
	RootDir  string
}

// GenerateTexts generates texts for all languages
func (t *Texts) GenerateTexts(client openai.Client, logger *slog.Logger) {
	t.Texts = make([]*Text, len(t.LangsOut))
	var wg sync.WaitGroup

	var j int
	for i, langOut := range t.LangsOut {
		if t.LangIn == langOut {
			text := NewText(t.RootDir, langOut, client, logger)
			err := text.GenerateText(nil, logger)
			if err != nil {
				logger.Error("Failed to generate text", "error", err)
				os.Exit(1)
			}
			t.Texts[i] = text
			j = i
			break
		}
	}

	for i, langOut := range t.LangsOut {
		if t.LangIn != langOut {
			wg.Add(1)
			go func(i int, langOut string) {
				defer wg.Done()
				t.Texts[i] = NewText(t.RootDir, langOut, client, logger)
				err := t.Texts[i].GenerateText(t.Texts[j], logger)
				if err != nil {
					logger.Error("Failed to generate text", "error", err)
					os.Exit(1)
				}
			}(i, langOut)
		}
	}

	wg.Wait()
}
