# GoCreator Technical Improvement Roadmap

## Overview

This document outlines technical improvements and feature expansions for GoCreator, focusing on implementation details, architecture, and development priorities.

---

## Current State

### Strengths
- ✅ Solid core functionality for video generation from slides and text
- ✅ Multi-language support with AI-powered translation and TTS
- ✅ Google Slides API integration
- ✅ Intelligent caching system to reduce API calls
- ✅ Video slides support (not just static images)
- ✅ Clean architecture with dependency injection and comprehensive tests

### Limitations
- ⚠️ CLI-only interface (no GUI)
- ⚠️ Limited input sources (local files, Google Slides only)
- ⚠️ Single output format (MP4 only)
- ⚠️ No subtitle/caption generation
- ⚠️ Manual configuration via CLI flags
- ⚠️ Limited customization options

---

## Feature Improvements

### 1. Ease of Use

#### Configuration File Support
**Priority**: High  
**Complexity**: Low

Replace CLI flags with YAML/JSON configuration files:

```yaml
# gocreator.yaml
version: "1.0"

project:
  name: "My Video Project"
  
input:
  source: google-slides  # local, google-slides, powerpoint, notion
  presentation_id: "1ABC-xyz123"
  language: en

output:
  languages: [en, fr, es, de, ja]
  format: mp4
  quality: high  # low, medium, high, ultra
  directory: ./output

voices:
  en: alloy
  fr: shimmer
  speed: 1.0

subtitles:
  enabled: true
  formats: [srt, vtt]
  burn_in: false

branding:
  intro_video: ./assets/intro.mp4
  outro_video: ./assets/outro.mp4
  watermark: ./assets/logo.png
```

**Implementation**:
```go
package config

type Config struct {
    Version   string
    Project   ProjectConfig
    Input     InputConfig
    Output    OutputConfig
    Voices    VoiceConfig
    Subtitles SubtitleConfig
    Branding  BrandingConfig
}

func Load(path string) (*Config, error)
func Validate(cfg *Config) error
```

**Timeline**: 2-3 weeks

---

#### Setup Wizard
**Priority**: Medium  
**Complexity**: Low

Interactive CLI wizard for initial setup:

```bash
gocreator setup
# Prompts for:
# - OpenAI API key
# - Google Slides credentials (optional)
# - Default languages
# - Voice preferences
# - Output directory
```

**Implementation**: Use libraries like `survey` or `promptui` for interactive prompts.

**Timeline**: 1-2 weeks

---

### 2. Input Sources

#### PowerPoint Support
**Priority**: High  
**Complexity**: High

Enable direct import of .pptx files with slides and speaker notes.

**Technical Approach**:
- Use `github.com/unidoc/unioffice` for PPTX parsing
- Or LibreOffice headless for slide rendering
- Extract speaker notes as narration text

**Architecture**:
```
powerpoint/
├── parser.go           # PPTX file parsing
├── slide_extractor.go  # Slide to image conversion
├── notes_extractor.go  # Speaker notes extraction
└── renderer.go        # Slide rendering
```

**Key Interfaces**:
```go
type PowerPointParser interface {
    Parse(path string) (*Presentation, error)
    ExtractSlides(pres *Presentation, outputDir string) ([]string, error)
    ExtractNotes(pres *Presentation) ([]string, error)
}
```

**Timeline**: 4-6 weeks

---

#### Markdown Support
**Priority**: Medium  
**Complexity**: Medium

Generate videos from Markdown documents with automatic slide creation.

**Technical Approach**:
- Parse Markdown with `goldmark` or `blackfriday`
- Generate slides from headers and content
- Support code syntax highlighting
- Embed images from Markdown

**Implementation**:
```go
package markdown

type MarkdownParser struct {
    fs     afero.Fs
    logger interfaces.Logger
}

func (p *MarkdownParser) Parse(path string) (*Document, error)
func (p *MarkdownParser) GenerateSlides(doc *Document) ([]Slide, error)
```

**Timeline**: 3-4 weeks

---

#### Notion Integration
**Priority**: Low  
**Complexity**: Medium

Import content from Notion pages.

**Technical Approach**:
- Use Notion API SDK
- Extract page blocks as slides
- Support embedded images and videos

**Timeline**: 3-4 weeks

---

### 3. Output Capabilities

#### Subtitle/Caption Generation
**Priority**: Very High  
**Complexity**: Medium

Automatic subtitle generation in multiple formats.

**Features**:
- SRT (SubRip) format
- VTT (WebVTT) format
- ASS (Advanced SubStation Alpha) for styling
- Burn-in subtitles (overlay on video)
- Word-level timing with forced alignment
- Customizable styling

**Architecture**:
```
subtitles/
├── generator.go      # Main subtitle generation
├── srt_writer.go     # SRT format
├── vtt_writer.go     # WebVTT format
├── ass_writer.go     # ASS format
├── synchronizer.go   # Audio-text sync
├── styles.go         # Styling config
└── burner.go        # Burn-in overlay
```

**Data Structures**:
```go
type Subtitle struct {
    Index     int
    StartTime time.Duration
    EndTime   time.Duration
    Text      string
    Styling   *SubtitleStyle
}

type SubtitleStyle struct {
    FontFamily   string
    FontSize     int
    FontColor    string
    BGColor      string
    Position     Position  // top, middle, bottom
    Alignment    Alignment // left, center, right
}
```

**Implementation Phases**:
1. Basic SRT generation (1-2 weeks)
2. VTT and ASS support (1 week)
3. Burn-in with ffmpeg (2 weeks)
4. Word-level timing with alignment tools (2-3 weeks)

**Timeline**: 6-8 weeks total

---

#### Audio-Only Export
**Priority**: Medium  
**Complexity**: Low

Export audio tracks without video for podcast/audiobook use.

**Formats**:
- MP3
- WAV
- M4A (with chapter markers)

**Implementation**:
```go
func (s *AudioService) ExportAudioOnly(
    audioPaths []string,
    metadata AudioMetadata,
    outputPath string,
    format AudioFormat,
) error
```

**Timeline**: 1-2 weeks

---

#### Social Media Format Presets
**Priority**: Medium  
**Complexity**: Low

Optimize output for different platforms with preset configurations.

**Presets**:
- YouTube: 16:9, 1920x1080, optimized bitrate
- Instagram: 9:16 vertical, 1080x1920
- TikTok: 9:16 vertical, optimized for mobile
- LinkedIn: 1:1 square or 16:9
- Twitter: 16:9, length optimization

**Configuration**:
```yaml
profiles:
  youtube:
    aspect_ratio: 16:9
    resolution: 1920x1080
    bitrate: 8M
    audio_bitrate: 192k
  
  instagram_reel:
    aspect_ratio: 9:16
    resolution: 1080x1920
    bitrate: 5M
    max_duration: 90s
```

**Timeline**: 2-3 weeks

---

### 4. Advanced Features

#### Background Music
**Priority**: Medium  
**Complexity**: Medium

Add background music with auto-ducking (lower volume during narration).

**Features**:
- Royalty-free music library integration
- Auto-ducking (reduce music volume during speech)
- Fade in/out
- Loop short tracks

**Implementation**:
```go
type BackgroundMusicService struct {
    fs     afero.Fs
    logger interfaces.Logger
}

func (s *BackgroundMusicService) AddBackgroundMusic(
    videoPath string,
    musicPath string,
    config MusicConfig,
) error
```

**Configuration**:
```yaml
background_music:
  enabled: true
  file: ./assets/music.mp3
  volume: 0.2
  fade_in: 2s
  fade_out: 3s
  ducking:
    enabled: true
    reduction: 0.5  # Reduce to 50% during speech
```

**Timeline**: 3-4 weeks

---

#### Quality Profiles
**Priority**: Medium  
**Complexity**: Low

Multiple quality presets for different use cases.

**Profiles**:
```yaml
profiles:
  ultra:
    resolution: 3840x2160  # 4K
    bitrate: 50M
    audio_bitrate: 320k
    
  high:
    resolution: 1920x1080  # Full HD
    bitrate: 8M
    audio_bitrate: 192k
    
  medium:
    resolution: 1280x720  # HD
    bitrate: 4M
    audio_bitrate: 128k
    
  low:
    resolution: 854x480  # SD
    bitrate: 1M
    audio_bitrate: 96k
```

**Timeline**: 1-2 weeks

---

#### Batch Processing
**Priority**: High  
**Complexity**: Medium

Process multiple projects in parallel with queue management.

**Features**:
- Process multiple presentations
- Job queue with priorities
- Progress tracking
- Concurrent processing with limits

**Implementation**:
```go
type BatchProcessor struct {
    queue      *JobQueue
    maxWorkers int
    logger     interfaces.Logger
}

func (p *BatchProcessor) AddJob(job *VideoJob) error
func (p *BatchProcessor) Start() error
func (p *BatchProcessor) GetStatus() BatchStatus
```

**CLI Usage**:
```bash
gocreator batch --manifest projects.yaml --parallel 4
```

**Timeline**: 3-4 weeks

---

### 5. User Interfaces

#### Desktop GUI Application
**Priority**: High  
**Complexity**: Very High

Native desktop application for non-technical users.

**Technology Choice**: Tauri (Rust + Web)
- Smaller bundle size than Electron
- Better performance
- Native system integration

**Features**:
- Drag-and-drop slide upload
- Visual timeline editor
- Real-time preview
- Settings panel
- Project management
- Export queue

**Architecture**:
```
desktop-app/
├── src/
│   ├── main.rs              # Tauri backend
│   ├── commands/            # Rust commands
│   │   ├── project.rs
│   │   ├── video.rs
│   │   └── settings.rs
│   └── frontend/            # React/Svelte
│       ├── components/
│       ├── pages/
│       └── store/
```

**Implementation Phases**:
1. Foundation & project structure (2-3 weeks)
2. Editor UI components (3-4 weeks)
3. Backend integration with CLI (3-4 weeks)
4. Polish & UX improvements (2-3 weeks)
5. Distribution & auto-update (2-3 weeks)

**Timeline**: 12-17 weeks

---

#### Web Application
**Priority**: High  
**Complexity**: Very High

Browser-based video editor with cloud capabilities.

**Technology Stack**:
- Frontend: React + TypeScript + Tailwind
- Backend: Go + Chi router
- Database: PostgreSQL
- Storage: S3-compatible (MinIO/AWS S3)
- Queue: Redis

**Architecture**:
```
web-app/
├── frontend/
│   └── src/
│       ├── components/
│       ├── pages/
│       └── services/
├── backend/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes.go
│   ├── services/
│   └── models/
└── docker-compose.yml
```

**Database Schema**:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMP
);

CREATE TABLE projects (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name VARCHAR(255),
    config JSONB,
    created_at TIMESTAMP
);

CREATE TABLE videos (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    language VARCHAR(10),
    status VARCHAR(50),
    s3_url VARCHAR(500)
);
```

**API Endpoints**:
```
POST   /api/v1/auth/login
GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/:id
POST   /api/v1/projects/:id/generate
GET    /api/v1/videos/:id
```

**Implementation Phases**:
1. Backend API (4-6 weeks)
2. Frontend foundation (3-4 weeks)
3. Editor implementation (4-6 weeks)
4. Video processing integration (3-4 weeks)
5. Deployment & infrastructure (2-3 weeks)

**Timeline**: 16-23 weeks

---

### 6. AI & Automation

#### Auto-Editing
**Priority**: Medium  
**Complexity**: Very High

Intelligent video editing with AI-powered features.

**Features**:
- Silence detection and removal
- Filler word removal ("um", "uh", "like")
- Pace optimization
- Auto-reframing for different aspect ratios

**Implementation**:
```go
package autoediting

type SilenceDetector struct {
    Threshold    float64  // dB level
    MinDuration  time.Duration
}

func (d *SilenceDetector) DetectSilence(audioPath string) ([]TimeRange, error)

type FillerWordDetector struct {
    Words []string
}

func (d *FillerWordDetector) Detect(transcript Transcript) ([]TimeRange, error)
```

**Configuration**:
```yaml
auto_editing:
  enabled: true
  remove_silence:
    enabled: true
    threshold: -40dB
    min_duration: 1s
  remove_filler_words:
    enabled: true
    words: ["um", "uh", "like", "you know"]
  pace_optimization:
    target_wpm: 150
    max_speedup: 1.3x
```

**Timeline**: 8-10 weeks

---

#### AI Slide Generation
**Priority**: Low  
**Complexity**: Very High

Generate slides automatically from text/topics.

**Features**:
- Generate slides from topic
- Automatic layout selection
- Stock image integration
- Consistent design theme

**CLI Usage**:
```bash
gocreator generate --topic "Introduction to Machine Learning" --slides 10
```

**Timeline**: 10-12 weeks

---

## Technical Architecture

### Plugin System

Enable extensibility without bloating the core.

**Structure**:
```
plugins/
├── input/
│   ├── powerpoint/
│   ├── notion/
│   └── markdown/
├── output/
│   ├── subtitles/
│   └── social-media/
└── processors/
    ├── background-music/
    └── transitions/
```

**Plugin Interface**:
```go
package plugins

type Plugin interface {
    Name() string
    Version() string
    Initialize(config map[string]interface{}) error
    Shutdown() error
}

type InputPlugin interface {
    Plugin
    Load(source string) (*Content, error)
    SupportedFormats() []string
}

type OutputPlugin interface {
    Plugin
    Export(video *Video, config ExportConfig) error
    SupportedFormats() []string
}
```

**Timeline**: 4-6 weeks

---

### System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                         │
├──────────────┬──────────────┬──────────────┬───────────┤
│     CLI      │   Desktop    │   Web App    │    API    │
└──────┬───────┴──────┬───────┴──────┬───────┴─────┬─────┘
       │              │              │             │
       └──────────────┴──────────────┴─────────────┘
                      │
       ┌──────────────▼──────────────┐
       │      Core Services           │
       │  • Project Manager           │
       │  • Video Creator             │
       │  • Config Manager            │
       │  • Plugin Registry           │
       └──────────────┬──────────────┘
                      │
       ┌──────────────▼──────────────┐
       │   Processing Services        │
       │  • Text Processor            │
       │  • Audio Generator           │
       │  • Video Generator           │
       │  • Translation Service       │
       │  • Subtitle Generator        │
       └──────────────┬──────────────┘
                      │
       ┌──────────────▼──────────────┐
       │  External Integrations       │
       │  • OpenAI API                │
       │  • Google APIs               │
       │  • Storage (S3/Local)        │
       └──────────────┴──────────────┘
```

---

## Implementation Priorities

### Phase 1: Core Improvements (2-3 months)
**Focus**: Essential features that provide immediate value

- [x] Configuration file support (3 weeks)
- [x] Subtitle/caption generation (6-8 weeks)
- [x] Setup wizard (2 weeks)
- [x] Quality profiles (2 weeks)
- [x] Background music (4 weeks)

**Total**: ~17-19 weeks

---

### Phase 2: Platform Expansion (3-4 months)
**Focus**: Broader input/output support

- [x] PowerPoint integration (6 weeks)
- [x] Markdown support (4 weeks)
- [x] Social media format presets (3 weeks)
- [x] Audio-only export (2 weeks)
- [x] Batch processing (4 weeks)

**Total**: ~19 weeks

---

### Phase 3: User Interfaces (5-6 months)
**Focus**: Accessibility for non-technical users

- [x] Desktop GUI application (17 weeks)
- [x] Web application MVP (23 weeks)
- [x] Plugin architecture (6 weeks)

**Total**: ~46 weeks (can be parallelized)

---

### Phase 4: Advanced AI (4-6 months)
**Focus**: Cutting-edge automation

- [x] Auto-editing features (10 weeks)
- [x] AI slide generation (12 weeks)
- [x] Voice cloning integration (8 weeks)
- [x] Content repurposing (6 weeks)

**Total**: ~36 weeks (can be parallelized)

---

## Dependency Graph

```
Configuration System (Foundation)
        │
        ├─> Subtitles
        ├─> PowerPoint
        ├─> Markdown
        └─> Setup Wizard
                │
                ├─> Desktop GUI
                └─> Web App
                        │
                        └─> Auto-Editing
                        └─> AI Features
```

---

## Development Guidelines

### Code Quality
1. Follow existing architecture patterns (dependency injection)
2. Write comprehensive unit tests for all new code
3. Use interfaces for external dependencies
4. Document public APIs and complex logic
5. Maintain backward compatibility

### Testing Strategy
1. Unit tests with mocked dependencies
2. Integration tests with test fixtures
3. Benchmarks for performance-critical paths
4. E2E tests for major workflows

### Performance Targets
- Video generation: <5 minutes for 10-minute video
- API response time: <200ms p95
- Memory usage: <2GB during generation
- Cache hit rate: >70%

---

## User Personas (Technical Focus)

### Persona 1: Course Creator
**Technical Need**: Bulk video generation from existing PowerPoint slides  
**Key Features**: PowerPoint import, batch processing, subtitles, multi-language  
**Workflow**: Import → Configure → Generate → Export  
**Volume**: 50-100 videos per course

### Persona 2: Developer Advocate
**Technical Need**: Automated video creation from documentation  
**Key Features**: Markdown import, code highlighting, API/CLI automation, CI/CD integration  
**Workflow**: Write docs → Auto-generate → Review → Publish  
**Integration**: GitHub Actions workflow

### Persona 3: Content Creator
**Technical Need**: Rapid video production with consistent quality  
**Key Features**: Templates, social media presets, batch processing, automation  
**Workflow**: Script → Generate → Optimize → Upload  
**Volume**: 3-5 videos per week

### Persona 4: Corporate Trainer
**Technical Need**: Multilingual training videos with compliance requirements  
**Key Features**: PowerPoint import, multi-language, subtitles, quality control  
**Workflow**: Update materials → Generate → Review → Deploy to LMS  
**Scale**: 10+ languages, 100+ videos

### Persona 5: Marketing Team
**Technical Need**: Quick iteration and A/B testing  
**Key Features**: Templates, social media formats, batch processing, version control  
**Workflow**: Design → Generate variants → Test → Deploy  
**Volume**: Multiple campaigns per month

---

## Competitive Technical Analysis

### vs. Synthesia
**Our Advantage**: Open-source, self-hosted, full customization, API-first, no vendor lock-in  
**Their Advantage**: AI avatars, mature UI  
**Focus**: Add desktop/web GUI, plugin system for avatars

### vs. Pictory
**Our Advantage**: Better slide support, Google Slides integration, local processing  
**Their Advantage**: AI-powered automation  
**Focus**: Add AI features (auto-editing, smart selection)

### vs. Descript
**Our Advantage**: Simpler workflow for slide-based videos, purpose-built  
**Their Advantage**: Advanced audio/video editing  
**Focus**: Maintain simplicity while adding power features

---

## Contributing

### For New Features
1. Open an issue with technical proposal
2. Discuss architecture and API design
3. Implement with tests
4. Submit PR with documentation
5. Update this roadmap

### Development Setup
```bash
# Clone repository
git clone https://github.com/Napolitain/gocreator.git
cd gocreator

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o gocreator ./cmd/gocreator
```

### Plugin Development
See `docs/PLUGIN_DEVELOPMENT.md` for creating custom plugins.

---

## Conclusion

This roadmap provides a clear technical path for evolving GoCreator from a CLI tool into a comprehensive video creation platform. The focus is on:

1. **Immediate Value**: Subtitles, config files, setup wizard
2. **Broad Compatibility**: PowerPoint, Markdown, multiple formats
3. **Accessibility**: Desktop and web interfaces
4. **Innovation**: AI-powered automation

**Next Steps**:
1. Start with Phase 1 features (config system, subtitles)
2. Gather community feedback on priorities
3. Accept contributions following the guidelines
4. Iterate based on user needs

---

**Document Version**: 1.0  
**Last Updated**: November 2025  
**Status**: Living document - updated based on community feedback
