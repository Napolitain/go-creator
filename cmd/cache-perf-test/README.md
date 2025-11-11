# Cache Performance Testing Tool

This tool helps measure cache hit rates and performance improvements from GoCreator's multi-layer caching system.

## Features

- **Mock API Calls**: Simulates OpenAI API calls with configurable 50ms delays instead of making real API calls
- **Cache Hit Tracking**: Tracks hits/misses for all cache layers:
  - Translation cache
  - Audio cache
  - Video segment cache
  - Final video cache
- **Performance Metrics**: Measures:
  - Total duration for each run
  - Cache hit rates
  - API call counts
  - Speedup from caching
- **Multiple Scenarios**: Tests different project sizes:
  - Small: 3 slides, 2 languages
  - Medium: 5 slides, 3 languages
  - Large: 10 slides, 4 languages

## Usage

### Run the performance test:

```bash
cd cmd/cache-perf-test
go run main.go
```

### Example Output:

```
=== GoCreator Cache Performance Testing Tool ===

### Testing Scenario: Small Project (3 slides, 2 languages) ###

  Run 1/3...
    Duration: 1.234s
    Cache Hit Rate: 0.0%
    API Calls (Translation): 6
    API Calls (TTS): 6
    Segment Cache Hits: 0
    Final Cache Hits: 0

  Run 2/3...
    Duration: 234ms
    Cache Hit Rate: 87.5%
    API Calls (Translation): 0
    API Calls (TTS): 0
    Segment Cache Hits: 6
    Final Cache Hits: 2

  Run 3/3...
    Duration: 189ms
    Cache Hit Rate: 100.0%
    API Calls (Translation): 0
    API Calls (TTS): 0
    Segment Cache Hits: 6
    Final Cache Hits: 2


=== Performance Summary ===

### Small Project (3 slides, 2 languages) ###
  First Run:  1.234s (0% cache hits, 12 API calls)
  Last Run:   189ms (100% cache hits, 0 API calls)
  Speedup:    6.5x faster
  API Savings: 12 calls (100% reduction)

=== Key Insights ===
- First run: No cache, all operations performed
- Subsequent runs: Cache utilized, significant speedup
- API calls are simulated with 50ms delay each
- Segment and final video caching provide the largest performance gains

Note: Actual performance will vary based on real API latency,
      FFmpeg encoding time, and disk I/O speed.
```

## How It Works

1. **Mock API Client**: Replaces OpenAI API calls with 50ms sleeps
   - `ChatCompletion()` - simulates translation API calls
   - `GenerateSpeech()` - simulates TTS API calls

2. **Cache Tracking Logger**: Monitors all cache operations
   - Detects cache hits from log messages
   - Counts cache misses when operations are performed
   - Tracks hits/misses for each cache layer

3. **Scenario Execution**: Runs each scenario multiple times
   - First run: Cold cache (no cached data)
   - Subsequent runs: Warm cache (reuses cached data)
   - Measures duration and cache performance

4. **Performance Metrics**: Calculates and reports:
   - Cache hit rates per run
   - Speedup from first to last run
   - API call reduction
   - Cache layer breakdown

## Customizing Scenarios

Edit the `scenarios` slice in `main()` to test different configurations:

```go
scenarios := []Scenario{
    {
        Name:              "Custom Scenario",
        NumSlides:         7,      // Number of slides
        NumLanguages:      5,      // Number of output languages
        TransitionEnabled: true,   // Enable video transitions
        RunCount:          4,      // Number of times to run
    },
}
```

## Customizing API Delays

Modify the delay values when creating the mock client in `runScenario()`:

```go
// Default: 50ms for both translation and TTS
mockClient := NewMockOpenAIClient(50*time.Millisecond, 50*time.Millisecond)

// Example: Simulate slower API responses
mockClient := NewMockOpenAIClient(200*time.Millisecond, 150*time.Millisecond)
```

## Understanding the Results

### Cache Hit Rate
- **0%**: First run, no cache available
- **50-80%**: Partial cache (some layers cached, others not)
- **100%**: Full cache (all operations cached)

### Speedup Factor
- **1x**: No improvement (first run)
- **2-5x**: Good caching (audio/translation cached)
- **5-10x**: Excellent caching (all layers cached)
- **10x+**: Outstanding (large projects with full cache)

### API Call Reduction
- Shows how many API calls were avoided due to caching
- Each API call costs money, so this represents direct cost savings
- 100% reduction on cached runs is ideal

## Notes

- This tool uses mock data and doesn't require FFmpeg or actual API keys
- Real-world performance will include FFmpeg encoding time (not simulated here)
- Disk I/O speed can affect cache read/write performance
- Network latency to OpenAI APIs is typically 100-500ms per call
- The 50ms delay is conservative; real API calls are often slower
