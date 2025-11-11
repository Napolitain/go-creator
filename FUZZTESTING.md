# Fuzz Testing Guide

This document explains how to use and write fuzz tests for GoCreator.

## What is Fuzz Testing?

Fuzz testing (or fuzzing) is an automated testing technique that provides random, unexpected, or invalid data as inputs to a program. Go has built-in support for fuzz testing since Go 1.18, which helps discover edge cases, security vulnerabilities, and crashes that traditional unit tests might miss.

## Why Fuzz Testing?

GoCreator handles various types of user inputs:
- Text files with custom delimiters
- Translation text in multiple languages
- Cache keys and file paths
- Unicode content and special characters

Fuzz testing helps ensure the code is robust against:
- Unexpected input formats
- Edge cases (empty strings, very long strings, special characters)
- Unicode and encoding issues
- Delimiter conflicts
- Line ending variations (Windows vs Unix)

## Running Fuzz Tests

### Basic Usage

Run a specific fuzz test:
```bash
go test -fuzz=FuzzTextService_Hash -fuzztime=30s ./internal/services/
```

Parameters:
- `-fuzz=FuzzTestName` - Name of the fuzz test to run
- `-fuzztime=30s` - How long to run (default: unlimited)
- Can also use `-fuzztime=1000x` to run for a specific number of iterations

### Running All Fuzz Tests

```bash
# Run each fuzz test for 10 seconds
go test -fuzz=FuzzTextService_Hash -fuzztime=10s ./internal/services/
go test -fuzz=FuzzTextService_LoadAndSave -fuzztime=10s ./internal/services/
go test -fuzz=FuzzTextService_SaveHashes -fuzztime=10s ./internal/services/
go test -fuzz=FuzzTranslationService_getCacheKey -fuzztime=10s ./internal/services/
go test -fuzz=FuzzTranslationService_MemoryCache -fuzztime=10s ./internal/services/
go test -fuzz=FuzzTranslationService_DiskCache -fuzztime=10s ./internal/services/
```

### When a Fuzz Test Finds a Failure

When a fuzz test discovers a failing input, it:
1. Minimizes the input to the smallest failing case
2. Saves it to `testdata/fuzz/FuzzTestName/`
3. Shows you how to re-run it

Example:
```bash
# The fuzz test found a failure and saved it
Failing input written to testdata/fuzz/FuzzTextService_Hash/abc123def456

# Re-run just that specific failing case
go test -run=FuzzTextService_Hash/abc123def456 ./internal/services/
```

## Available Fuzz Tests

### TextService Tests (`text_fuzz_test.go`)

#### FuzzTextService_Hash
Tests the SHA256 hash function with random inputs.

**What it tests:**
- Hash is deterministic (same input → same output)
- Hash is always 64 characters (SHA256 hex)
- Different inputs produce different hashes
- Never panics on any input

**Example:**
```bash
go test -fuzz=FuzzTextService_Hash -fuzztime=30s ./internal/services/
```

#### FuzzTextService_LoadAndSave
Tests text file parsing and saving with various inputs.

**What it tests:**
- Loading text files with various delimiters and content
- Round-trip consistency (save → load → same result)
- Handling of multiline texts
- Edge cases with empty strings and whitespace
- Line ending normalization (\r\n vs \n)

**Known limitations discovered:**
- Text that is exactly "-" cannot be round-tripped (it's the delimiter)
- Texts containing a line that's exactly "-" will be split on reload
- Whitespace-only texts become empty after trimming

**Example:**
```bash
go test -fuzz=FuzzTextService_LoadAndSave -fuzztime=30s ./internal/services/
```

#### FuzzTextService_SaveHashes
Tests hash file operations with random hash values.

**What it tests:**
- Saving and loading arrays of hashes
- Round-trip consistency
- Handling of various string formats

**Known limitations:**
- Hash values cannot contain newlines (newlines are used as delimiters)

**Example:**
```bash
go test -fuzz=FuzzTextService_SaveHashes -fuzztime=30s ./internal/services/
```

### TranslationService Tests (`translation_fuzz_test.go`)

#### FuzzTranslationService_getCacheKey
Tests cache key generation with random text and language inputs.

**What it tests:**
- Cache key is deterministic
- Cache key is always 64 characters (SHA256 hex)
- Different inputs produce different cache keys
- Never panics on any input

**Example:**
```bash
go test -fuzz=FuzzTranslationService_getCacheKey -fuzztime=30s ./internal/services/
```

#### FuzzTranslationService_MemoryCache
Tests memory cache operations with random keys and values.

**What it tests:**
- Setting and getting values from memory cache
- Cache consistency
- Handling of various string formats as keys and values

**Example:**
```bash
go test -fuzz=FuzzTranslationService_MemoryCache -fuzztime=30s ./internal/services/
```

#### FuzzTranslationService_DiskCache
Tests disk cache operations with random keys and values.

**What it tests:**
- Writing and reading cache files
- Handling of special characters in cache keys
- Cache consistency across disk operations

**Note:** Some keys may not be valid filenames (e.g., those with path separators), so the test skips those cases.

**Example:**
```bash
go test -fuzz=FuzzTranslationService_DiskCache -fuzztime=30s ./internal/services/
```

## Writing New Fuzz Tests

### Basic Structure

```go
func FuzzMyFunction(f *testing.F) {
    // 1. Add seed corpus - interesting test cases
    f.Add("normal input")
    f.Add("")  // edge case: empty
    f.Add("unicode: 你好🌍")  // edge case: unicode
    
    // 2. Fuzz function - will be called with random inputs
    f.Fuzz(func(t *testing.T, input string) {
        // Your test code here
        result := MyFunction(input)
        
        // Check invariants - properties that should always be true
        if result == "" {
            t.Error("Result should never be empty")
        }
    })
}
```

### Best Practices

1. **Add Good Seed Corpus**: Start with interesting edge cases
   - Empty strings
   - Very long strings
   - Unicode characters
   - Special characters relevant to your function
   - Known problematic inputs

2. **Test Invariants**: Check properties that should always hold
   - Function never panics
   - Output format is correct
   - Deterministic behavior (same input → same output)
   - Consistency across operations

3. **Handle Known Limitations**: Document and skip cases that can't work
   ```go
   if containsProblematicPattern(input) {
       t.Skip("Input contains pattern that can't be handled")
   }
   ```

4. **Normalize Comparisons**: When comparing outputs, normalize things like:
   - Line endings (\r\n vs \n)
   - Whitespace
   - Unicode normalization

### Example: Adding a New Fuzz Test

Let's say you want to add a fuzz test for a hypothetical `SanitizeFilename` function:

```go
func FuzzSanitizeFilename(f *testing.F) {
    // Seed corpus with interesting cases
    f.Add("normal.txt")
    f.Add("../../etc/passwd")  // path traversal
    f.Add("file\x00name.txt")  // null byte
    f.Add("unicode_文件.txt")  // unicode
    f.Add("very" + strings.Repeat("long", 100) + ".txt")  // long name
    
    f.Fuzz(func(t *testing.T, filename string) {
        // Function should never panic
        result := SanitizeFilename(filename)
        
        // Result should not contain path separators
        if strings.Contains(result, "/") || strings.Contains(result, "\\") {
            t.Errorf("Sanitized filename contains path separator: %q", result)
        }
        
        // Result should not contain null bytes
        if strings.Contains(result, "\x00") {
            t.Errorf("Sanitized filename contains null byte: %q", result)
        }
        
        // Result should not be empty for non-empty input
        if filename != "" && result == "" {
            t.Errorf("Empty result for non-empty input: %q", filename)
        }
    })
}
```

## Continuous Integration

Fuzz tests are **not** run in CI by default because:
1. They can run indefinitely
2. They generate random failures that may not be reproducible

However, CI does run any saved failing test cases from `testdata/fuzz/` as regular unit tests.

To add fuzz testing to CI:
1. Run fuzz tests locally for a reasonable time (e.g., 5 minutes per test)
2. Commit any discovered failing cases to `testdata/fuzz/`
3. Fix the issues or document known limitations
4. The failing cases will be tested in CI as regression tests

## Tips and Tricks

1. **Start with short fuzz times** (10-30 seconds) to see if basic fuzzing works
2. **Gradually increase time** for longer, more thorough testing
3. **Look at coverage**: Fuzzing reports "new interesting" cases that increase code coverage
4. **Parallelize**: You can run multiple fuzz tests in parallel in different terminals
5. **Save corpus**: The `testdata/fuzz/` directory contains all interesting cases found

## Further Reading

- [Go Fuzzing Documentation](https://go.dev/doc/fuzz/)
- [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)
- [Go Blog: Fuzzing is Beta Ready](https://go.dev/blog/fuzz-beta)

## Contributing

When adding new code to GoCreator:

1. Consider adding fuzz tests for:
   - Functions that parse user input
   - Functions that handle strings or files
   - Functions that generate hashes or keys
   - Functions with complex logic or edge cases

2. Run fuzz tests before submitting PRs:
   ```bash
   go test -fuzz=FuzzYourNewTest -fuzztime=60s ./internal/services/
   ```

3. Document any known limitations discovered during fuzzing

4. Commit any failing test cases found during development
