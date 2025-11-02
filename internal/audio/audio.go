package audio

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gocreator/internal/text"

	"github.com/openai/openai-go/v3"
)

// Model represents the model to use for audio generation
type Model string

const (
	// Audio generation models
	OpenAITTS Model = "openai"
	GoogleTTS Model = "google"
)

// GenerateAudioOpenAI generates audio using OpenAI TTS
func GenerateAudioOpenAI(client openai.Client, textContent, speechFilePath, cachedHash string, logger *slog.Logger) error {
	newHash := fmt.Sprintf("%x", sha256.Sum256([]byte(textContent)))
	if newHash == cachedHash {
		return nil
	}

	response, err := client.Audio.Speech.New(
		context.Background(),
		openai.AudioSpeechNewParams{
			Model:          openai.SpeechModelTTS1HD,
			Input:          textContent,
			Voice:          openai.AudioSpeechNewParamsVoice("onyx"),
			ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
		},
	)
	if err != nil {
		return err
	}

	file, err := os.Create(speechFilePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, response.Body)
	if err != nil {
		file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

// GenerateAudios generates audio for all texts
func GenerateAudios(texts *text.Texts, client openai.Client, audioModel Model, logger *slog.Logger) map[string][]string {
	audiosLangToPath := make(map[string][]string)
	var wgTop sync.WaitGroup
	for _, txt := range texts.Texts {
		wgTop.Add(1)
		go func(txt *text.Text) {
			defer wgTop.Done()
			audioDir := filepath.Join(txt.DataDir, "cache", txt.Lang, "audio")
			if _, err := os.Stat(audioDir); os.IsNotExist(err) {
				err := os.Mkdir(audioDir, os.ModePerm)
				if err != nil {
					logger.Error("Failed to create audio directory", "error", err)
					os.Exit(1)
				}
			}
			cachedHashes := txt.GenerateCacheHashes(audioDir)
			currentHashes := txt.Hashes
			hashFile, err := os.Create(filepath.Join(audioDir, "hashes"))
			if err != nil {
				logger.Error("Failed to create hash file", "dir", audioDir, "error", err)
				os.Exit(1)
			}
			defer hashFile.Close()
			writer := bufio.NewWriter(hashFile)

			var wg sync.WaitGroup
			results := make([]string, len(currentHashes))

			for j, currentHash := range currentHashes {
				wg.Add(1)
				var cachedHash string
				if j < len(cachedHashes) {
					cachedHash = cachedHashes[j]
				} else {
					cachedHash = ""
				}
				writeString, err := writer.WriteString(currentHash + "\n")
				if err != nil {
					logger.Error("Failed to write hash", "bytes", writeString, "error", err)
					os.Exit(1)
				}
				go func(j int, currentHash, cachedHash string) {
					defer wg.Done()
					audioPath := filepath.Join(audioDir, fmt.Sprintf("%d.mp3", j))
					if currentHash == cachedHash {
						results[j] = audioPath
						return
					}

					var err error
					if audioModel == OpenAITTS {
						err = GenerateAudioOpenAI(client, txt.SlidesText[j].Text, audioPath, cachedHash, logger)
					} else if audioModel == GoogleTTS {
						logger.Info("Google TTS not implemented, skipping")
						return
					} else {
						logger.Error("Unsupported audio model", "model", audioModel)
						os.Exit(1)
					}

					if err != nil {
						logger.Error("Failed to generate audio", "error", err)
						os.Exit(1)
					}
					results[j] = audioPath
				}(j, currentHash, cachedHash)
			}

			wg.Wait()
			err = writer.Flush()
			if err != nil {
				logger.Error("Failed to flush writer", "error", err)
				os.Exit(1)
			}
			audiosLangToPath[txt.Lang] = results
		}(txt)

	}
	wgTop.Wait()
	return audiosLangToPath
}
