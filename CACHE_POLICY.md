# Cache Management Policy

This document describes the caching strategy used in gocreator to optimize performance and reduce API costs.

## Overview

The application implements a multi-layered caching strategy to avoid redundant API calls and ffmpeg operations:

1. **Translation Cache** - File-based cache for translated texts
2. **Audio Generation Cache** - File-based cache with hash validation for generated audio
3. **Video Segment Cache** - Hash-based cache for individual video segments (NEW)
4. **Final Video Cache** - Hash-based cache for concatenated final videos (NEW)
5. **In-Memory Cache** - General-purpose cache service for runtime data

## 1. Translation Cache

**Location**: `data/cache/{language}/text/texts.txt`

**Purpose**: Avoid re-translating the same texts when re-running the tool

**Strategy**:
- When translating to a target language, the service first checks if a cached translation file exists
- If found, it loads the cached translations instead of calling the OpenAI API
- If not found, it translates the texts and saves them to the cache file

**Cache Key**: Language code (e.g., "es", "fr", "de")

**Expiration**: **Never expires** - Filesystem cache persists indefinitely

**Invalidation**: 
- Manual deletion of cache files
- When source texts change (detected by content comparison)

## 2. Audio Generation Cache

**Location**: `data/cache/{language}/audio/{index}.mp3` and corresponding `.hash` files

**Purpose**: Avoid regenerating audio for the same text content

**Strategy**:
- Before generating audio, the service computes a SHA256 hash of the text
- It checks if an audio file with a matching hash already exists
- If cached, it reuses the existing file
- If not, it generates new audio and saves both the audio file and its hash

**Cache Key**: SHA256 hash of the input text

**Hash Files**: Each audio file has a corresponding `.hash` file containing the SHA256 hash of the text that generated it

**Expiration**: **Never expires** - Filesystem cache persists indefinitely

**Invalidation**: 
- Automatic when text content changes (hash mismatch detected via SHA256)
- Manual deletion of audio files or hash files

## 3. Video Segment Cache

**Location**: `data/out/.temp/video_{index}.mp4` and corresponding `.hash` files

**Purpose**: Cache individual video segments to avoid re-encoding unchanged content

**Strategy**:
- Before generating a segment, computes a SHA256 hash of: slide content + audio content + dimensions
- If a cached segment with matching hash exists, it's reused
- If not cached, generates the segment and saves both the video file and its hash
- Each slide+audio combination is rendered as a separate video segment
- Segments are generated in parallel for performance
- All segments are then concatenated into the final video

**Cache Key**: SHA256 hash of (slide file + audio file + target dimensions)

**Hash Files**: Each video segment has a corresponding `.hash` file containing the SHA256 hash of its inputs

**Expiration**: **Never expires** - Filesystem cache persists indefinitely

**Invalidation**: 
- Automatic when slide content, audio content, or dimensions change (hash mismatch detected via SHA256)
- Manual deletion of segment files or hash files

**Benefits**:
- **Major performance gain**: Skips FFmpeg encoding for unchanged slide+audio pairs
- Enables parallel processing of segments
- Allows reuse of segments across runs if inputs haven't changed
- Simplifies debugging and inspection of individual segments
- No time-based expiration - only hash-based invalidation

## 4. Final Video Cache

**Location**: `data/out/output-{language}.mp4` and corresponding `.hash` files

**Purpose**: Cache the final concatenated video to avoid re-concatenation when inputs haven't changed

**Strategy**:
- Before concatenating segments, computes a SHA256 hash of: all segment files + transition type + transition duration
- If a cached final video with matching hash exists, it's reused
- If not cached, concatenates the segments and saves both the final video and its hash
- Transition-aware: different transition configurations produce different cache keys

**Cache Key**: SHA256 hash of (all video segments + transition type + transition duration)

**Hash Files**: Each final video has a corresponding `.hash` file containing the SHA256 hash of its inputs

**Expiration**: **Never expires** - Filesystem cache persists indefinitely

**Invalidation**: 
- Automatic when any segment changes (hash mismatch detected via SHA256)
- Automatic when transition configuration changes
- Manual deletion of final video files or hash files

**Benefits**:
- **Performance gain**: Skips FFmpeg concatenation when segments and transitions are unchanged
- Instant regeneration if inputs haven't changed
- Works with both simple concatenation and transition effects (xfade)
- Respects transition configuration changes

## 5. In-Memory Cache Service

**Implementation**: `internal/services/cache.go`

**Purpose**: General-purpose caching for runtime data that doesn't need persistence

**Features**:
- Time-based expiration (TTL)
- Automatic cleanup of expired entries
- Type-agnostic storage (stores interface{})

**Expiration**: **TTL-based** - Entries expire after configured duration

**Usage**: Can be used by services for temporary runtime caching needs

**Important**: This is the ONLY cache type with TTL expiration. All filesystem-based caches (translations, audio, video segments, final videos) persist indefinitely and use hash-based invalidation instead of time-based expiration.

## Cache Directory Structure

```
data/
├── cache/
│   ├── en/                    # English (input language)
│   │   ├── text/
│   │   │   └── texts.txt      # Not cached (same as input)
│   │   └── audio/
│   │       ├── 0.mp3
│   │       ├── 0.mp3.hash
│   │       ├── 1.mp3
│   │       ├── 1.mp3.hash
│   │       └── hashes         # Index of all hashes
│   ├── es/                    # Spanish translation
│   │   ├── text/
│   │   │   └── texts.txt      # Cached translation
│   │   └── audio/
│   │       ├── 0.mp3
│   │       ├── 0.mp3.hash
│   │       └── ...
│   └── fr/                    # French translation
│       └── ...
└── out/
    ├── .temp/                 # Video segment cache
    │   ├── video_0.mp4
    │   ├── video_0.mp4.hash   # NEW: Segment hash
    │   ├── video_1.mp4
    │   ├── video_1.mp4.hash   # NEW: Segment hash
    │   └── ...
    ├── output-en.mp4          # Final videos
    ├── output-en.mp4.hash     # NEW: Final video hash
    ├── output-es.mp4
    ├── output-es.mp4.hash     # NEW: Final video hash
    ├── output-fr.mp4
    └── output-fr.mp4.hash     # NEW: Final video hash
```

## Performance Implications

### First Run
- All translations are performed via OpenAI API
- All audio is generated via OpenAI TTS API
- All video segments are rendered via ffmpeg
- Final video is concatenated via ffmpeg

### Subsequent Runs (same texts and configuration)
- ✅ Translations are loaded from cache (no API calls)
- ✅ Audio files are reused (no API calls)
- ✅ **Video segments are reused (no ffmpeg encoding)** - NEW
- ✅ **Final video is reused (no ffmpeg concatenation)** - NEW
- **Result**: Near-instant video regeneration!

### Subsequent Runs (modified texts)
- ⚠️ Modified texts are re-translated (API calls for changed items)
- ⚠️ Modified audio is regenerated (API calls for changed items)
- ⚠️ **Affected video segments are re-encoded** - NEW
- ✅ Unchanged texts/audio/segments are reused from cache
- ⚠️ **Final video is re-concatenated** (hash changes due to segment changes) - NEW

### Transition Configuration Changes
- ✅ All translations remain cached
- ✅ All audio files remain cached
- ✅ All video segments remain cached
- ⚠️ **Final video is re-concatenated** (different transition configuration) - NEW
- **Result**: Only concatenation step re-runs, much faster than full regeneration

## Best Practices

1. **Preserve Cache Directory**: Keep the `data/cache` directory between runs to benefit from caching

2. **Preserve Temp Directory**: Keep the `data/out/.temp/` directory to benefit from video segment caching

3. **Monitor Hash Files**: Ensure `.hash` files are preserved alongside `.mp3`, `.mp4`, and final video files

4. **Clean Old Caches**: Periodically clean up cache directories for languages you no longer need

5. **Version Control**: Add `data/cache/`, `data/out/.temp/`, and `data/out/*.hash` to `.gitignore`

6. **Disk Space**: Video segment and final video caching requires disk space; monitor usage and clean old caches if needed

## Future Enhancements

Potential improvements to the caching system:

1. ~~**FFmpeg Output Caching**: Cache rendered video segments based on slide+audio hash~~ - ✅ **IMPLEMENTED**
2. ~~**Final Video Caching**: Cache concatenated final videos~~ - ✅ **IMPLEMENTED**
3. **Cache Compression**: Compress cache files to save disk space
4. **Cache Eviction Policy**: Implement LRU eviction for disk caches
5. **Cache Statistics**: Add logging for cache hit/miss rates
6. **Distributed Cache**: Support for shared cache across multiple machines
7. **Dimension-Based Segment Indexing**: Store segments with metadata for cross-project reuse
