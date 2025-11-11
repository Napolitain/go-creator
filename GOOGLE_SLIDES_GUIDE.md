# Google Slides Integration Guide

This guide explains how to set up and use Google Slides API with GoCreator.

## Overview

GoCreator can fetch slides and speaker notes directly from Google Slides presentations. This eliminates the need to manually export slides as images and allows you to keep your content in sync with your Google Slides presentation.

## Features

- **Automatic Slide Export**: Slides are automatically downloaded as images
- **Speaker Notes as Narration**: Speaker notes from each slide become the narration text
- **Live Updates**: Re-run to fetch the latest version of your presentation
- **Caching**: Downloaded slides and generated videos are cached for efficiency
- **Multiple Authentication Methods**: Support for both OAuth 2.0 (user authorization) and service account credentials

## Authentication Methods

GoCreator supports two authentication methods for Google Slides API:

### 🔐 OAuth 2.0 (Recommended for Personal Use)
- **Best for**: Individual users accessing their own presentations
- **Benefits**: Simple setup, uses your Google account, automatic token refresh
- **Access**: Only presentations owned by or shared with your Google account
- **Setup**: See [Option A: OAuth 2.0 Setup](#option-a-oauth-20-setup-recommended-for-personal-use) below

### 🔑 Service Account (Recommended for CI/CD)
- **Best for**: Automated workflows, servers, CI/CD pipelines
- **Benefits**: No user interaction needed, works headless
- **Access**: Only presentations explicitly shared with the service account email
- **Setup**: See [Option B: Service Account Setup](#option-b-service-account-setup-recommended-for-cicd) below

## Setup Instructions

### Common Setup Steps (Required for Both Methods)

#### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Create Project" or select an existing project
3. Note your Project ID

#### 2. Enable Google Slides API

1. In the Google Cloud Console, navigate to "APIs & Services" > "Library"
2. Search for "Google Slides API"
3. Click on it and click "Enable"
4. Also enable "Google Drive API" (optional but recommended for file access)

---

### Option A: OAuth 2.0 Setup (Recommended for Personal Use)

#### 1. Create OAuth 2.0 Credentials

1. Navigate to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth client ID"
3. If prompted, configure the OAuth consent screen:
   - Choose "External" user type
   - Fill in app name (e.g., "GoCreator")
   - Add your email as developer contact
   - Add scopes: `https://www.googleapis.com/auth/presentations` and `https://www.googleapis.com/auth/drive.file`
   - Add your email as a test user
   - Save and continue
4. Back at "Create OAuth client ID":
   - Choose "Desktop app" as application type
   - Enter a name (e.g., "GoCreator Desktop")
   - Click "Create"
5. Download the credentials JSON file
6. Save it as `~/.config/gocreator/oauth-credentials.json`

#### 2. Set Environment Variable

Set the `GOOGLE_OAUTH_CREDENTIALS` environment variable:

**Linux/macOS:**
```bash
export GOOGLE_OAUTH_CREDENTIALS="$HOME/.config/gocreator/oauth-credentials.json"
```

**Windows (PowerShell):**
```powershell
$env:GOOGLE_OAUTH_CREDENTIALS="C:\Users\YourName\.config\gocreator\oauth-credentials.json"
```

**Persistent setup (add to ~/.bashrc or ~/.zshrc):**
```bash
echo 'export GOOGLE_OAUTH_CREDENTIALS="$HOME/.config/gocreator/oauth-credentials.json"' >> ~/.bashrc
source ~/.bashrc
```

#### 3. First Run - Authorization

On your first run, GoCreator will:
1. Display an authorization URL
2. Ask you to visit the URL in your browser
3. Prompt you to sign in with your Google account
4. Ask you to authorize the application
5. Provide an authorization code
6. Request you to paste the code back into the terminal

The access token will be automatically saved and refreshed for future runs.

---

### Option B: Service Account Setup (Recommended for CI/CD)

#### 1. Create Service Account Credentials

1. Navigate to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "Service Account"
3. Enter a name (e.g., "gocreator-service")
4. Click "Create and Continue"
5. Skip granting access (optional)
6. Click "Done"
7. Click on the created service account
8. Go to "Keys" tab
9. Click "Add Key" > "Create new key"
10. Choose "JSON" format
11. Download the credentials file
12. Save it securely (e.g., `~/.config/gocreator/service-account-credentials.json`)

#### 2. Share Your Google Slides Presentation

**Important**: You must share your Google Slides presentation with the service account email address.

1. Open your Google Slides presentation
2. Click the "Share" button
3. Add the service account email (found in the credentials JSON file, looks like `gocreator-service@your-project.iam.gserviceaccount.com`)
4. Give it "Viewer" permission
5. Click "Send"

#### 3. Set Environment Variable

Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable:

**Linux/macOS:**
```bash
export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/gocreator/service-account-credentials.json"
```

**Windows (PowerShell):**
```powershell
$env:GOOGLE_APPLICATION_CREDENTIALS="C:\Users\YourName\.config\gocreator\service-account-credentials.json"
```

**Persistent setup (add to ~/.bashrc or ~/.zshrc):**
```bash
echo 'export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/gocreator/service-account-credentials.json"' >> ~/.bashrc
source ~/.bashrc
```

## Choosing the Right Authentication Method

| Feature | OAuth 2.0 | Service Account |
|---------|-----------|-----------------|
| **Setup Complexity** | Simple | Moderate |
| **User Interaction** | Required on first run | None needed |
| **Access Scope** | All presentations you own/have access to | Only presentations explicitly shared |
| **Token Management** | Automatic refresh | N/A |
| **Best For** | Personal development, local testing | CI/CD, automated workflows, servers |
| **Headless Support** | ❌ (requires browser) | ✅ |
| **Security** | User's Google account | Service account with limited access |

### Quick Recommendation

- **Use OAuth 2.0 if**: You're developing locally and accessing your own presentations
- **Use Service Account if**: You're running in CI/CD, on a server, or need headless operation

## Usage

### Find Your Presentation ID

The presentation ID is in the URL of your Google Slides:

```
https://docs.google.com/presentation/d/1ABC-xyz123_EXAMPLE-ID/edit
                                    └─────────────┬─────────────┘
                                            Presentation ID
```

### Run GoCreator

**With OAuth 2.0:**
```bash
export GOOGLE_OAUTH_CREDENTIALS="$HOME/.config/gocreator/oauth-credentials.json"
gocreator create --google-slides 1ABC-xyz123_EXAMPLE-ID --lang en --langs-out en,fr,es,de
```

**With Service Account:**

```bash
export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/gocreator/service-account-credentials.json"
gocreator create --google-slides 1ABC-xyz123_EXAMPLE-ID --lang en --langs-out en,fr,es,de
```

**Note**: GoCreator automatically detects which authentication method to use based on which environment variable is set. If both are set, OAuth 2.0 takes precedence.

### What Happens

1. GoCreator connects to Google Slides API using your credentials
2. Fetches the presentation metadata
3. Downloads each slide as a PNG image
4. Extracts speaker notes from each slide
5. Uses the notes as narration text for video generation
6. Generates videos with audio in all specified languages

## Best Practices

### 1. Speaker Notes

- **Be descriptive**: Write clear, complete sentences in speaker notes
- **One slide, one topic**: Keep each slide's narration focused
- **Timing**: Aim for 30-60 seconds of narration per slide
- **Language**: Write notes in your input language (specified with `--lang`)

### 2. Slide Design

- **Keep it simple**: Simple slides work best for video
- **High contrast**: Ensure text is readable
- **Avoid animations**: Slides are exported as static images
- **Consistent size**: Use a standard slide size (16:9 recommended)

### 3. Workflow

```bash
# 1. Create/update your Google Slides presentation
# 2. Add speaker notes to each slide
# 3. Run gocreator
gocreator create --google-slides YOUR_ID --lang en --langs-out en,fr,es

# 4. Check output in data/out/
ls data/out/
# output-en.mp4  output-fr.mp4  output-es.mp4
```

### 4. Caching

- Slides are cached in `data/slides/`
- Notes are saved to `data/texts.txt`
- To refresh from Google Slides, delete these and re-run
- Translations and audio are cached separately

## Troubleshooting

### Error: "no Google credentials found"

**Solution**: Set one of the environment variables:
- For OAuth 2.0: `export GOOGLE_OAUTH_CREDENTIALS="path/to/oauth-credentials.json"`
- For Service Account: `export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account-credentials.json"`

### Error: "failed to get presentation: googleapi: Error 404"

**Possible causes**:
1. Incorrect presentation ID
2. **OAuth 2.0**: Presentation not owned by or shared with your Google account
3. **Service Account**: Presentation not shared with service account email
4. Presentation deleted

**Solution**: 
- Verify the presentation ID
- **OAuth 2.0**: Ensure you have access to the presentation with your Google account
- **Service Account**: Check that you shared the presentation with the service account email
- Ensure the presentation exists

### Error: "failed to get presentation: googleapi: Error 403" (Permission Denied)

**Possible causes**:
1. **OAuth 2.0**: User hasn't authorized the application yet
2. **Service Account**: Service account doesn't have permission to access the presentation
3. API not enabled

**Solution**:
- **OAuth 2.0**: Run GoCreator and complete the authorization flow
- **Service Account**: Share the presentation with the service account email
- Ensure Google Slides API is enabled in your Google Cloud project

### Error: "failed to parse OAuth credentials" or "invalid credentials file"

**Possible causes**:
1. Wrong credentials file format
2. Using OAuth credentials file for service account (or vice versa)
3. Corrupted JSON file

**Solution**:
- Ensure you're using the correct environment variable for your credential type:
  - `GOOGLE_OAUTH_CREDENTIALS` for OAuth 2.0 desktop app credentials
  - `GOOGLE_APPLICATION_CREDENTIALS` for service account credentials
- Re-download the credentials file from Google Cloud Console
- Verify the JSON file is not corrupted

### OAuth 2.0 Authorization Issues

**Problem**: Browser doesn't open or authorization code is not accepted

**Solution**:
1. Manually copy and paste the authorization URL into your browser
2. Complete the authorization in the browser
3. Copy the authorization code from the browser
4. Paste it into the GoCreator terminal prompt
5. If the code is rejected, ensure you're copying the entire code without extra spaces

### Token Refresh Issues

**Problem**: "token refresh failed" or authentication errors after initial setup

**Solution**:
1. Delete the cached token: `rm ~/.config/gocreator/.gocreator-token.json`
2. Run GoCreator again to re-authorize
3. If issues persist, verify your OAuth credentials are still valid in Google Cloud Console

### Error: "failed to get thumbnail"

**Possible causes**:
1. Google Slides API not enabled
2. API quota exceeded
3. Network issues

**Solution**:
- Ensure Google Slides API is enabled in your project
- Check API usage quotas in Google Cloud Console
- Verify network connectivity

## Advanced Usage

### CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Generate Videos
on:
  push:
    branches: [main]

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup credentials
        env:
          GOOGLE_CREDENTIALS: ${{ secrets.GOOGLE_CREDENTIALS }}
        run: |
          echo "$GOOGLE_CREDENTIALS" > credentials.json
          echo "GOOGLE_APPLICATION_CREDENTIALS=$PWD/credentials.json" >> $GITHUB_ENV
      
      - name: Install gocreator
        run: |
          # Download and install gocreator
          
      - name: Generate videos
        run: |
          gocreator create --google-slides ${{ secrets.PRESENTATION_ID }} --lang en --langs-out en,fr
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: videos
          path: data/out/*.mp4
```

### Multiple Presentations

Process multiple presentations by running gocreator multiple times with different presentation IDs:

```bash
for id in "PRES_ID_1" "PRES_ID_2" "PRES_ID_3"; do
  gocreator create --google-slides "$id" --lang en --langs-out en,fr
  mv data/out/*.mp4 "output/$id/"
done
```

## Security Notes

1. **Keep credentials secure**: Never commit credentials to version control
2. **Use .gitignore**: Add `credentials.json` to your `.gitignore`
3. **Limit permissions**: Service account only needs "Viewer" access
4. **Rotate credentials**: Periodically create new service account keys
5. **Use secrets managers**: In production, use Google Secret Manager or similar

## Support

If you encounter issues:

1. Check this guide's troubleshooting section
2. Verify your Google Cloud project setup
3. Check API quotas and limits
4. Open an issue on GitHub with:
   - Error message (remove sensitive info)
   - Steps to reproduce
   - GoCreator version
   - OS and environment details
