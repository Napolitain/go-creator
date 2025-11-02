package video

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GetImageDimensions returns the width and height of an image
func GetImageDimensions(imagePath string) (int, int, error) {
	cmd := exec.Command("ffmpeg", "-i", imagePath, "-vf", "scale", "-vframes", "1", "-f", "null", "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("error running ffmpeg command: %v", err)
	}

	outputStr := string(output)
	re := regexp.MustCompile(`(\d+)x(\d+)`)
	matches := re.FindStringSubmatch(outputStr)
	if len(matches) < 3 {
		return 0, 0, fmt.Errorf("error parsing dimensions from ffmpeg output")
	}

	var width, height int
	fmt.Sscanf(matches[0], "%dx%d", &width, &height)

	return width, height, nil
}

// BuildFFmpegConcatCommand builds an ffmpeg command to concatenate multiple video files
func BuildFFmpegConcatCommand(videoFiles []string, finalOutput string) *exec.Cmd {
	args := []string{"-y"}

	for _, video := range videoFiles {
		args = append(args, "-i", video)
	}

	filterComplex := strings.Builder{}
	for i := range videoFiles {
		filterComplex.WriteString(fmt.Sprintf("[%d:v][%d:a]", i, i))
	}
	filterComplex.WriteString(fmt.Sprintf("concat=n=%d:v=1:a=1[outv][outa]", len(videoFiles)))

	args = append(args, "-filter_complex", filterComplex.String())
	args = append(args, "-map", "[outv]", "-map", "[outa]", finalOutput)

	return exec.Command("ffmpeg", args...)
}

// GenerateVideos generates videos for each slide/audio pair for all languages
func GenerateVideos(slides []string, audiosLangToPath map[string][]string, dataDir string, logger *slog.Logger) {
	firstSlide := slides[0]
	width, height, err := GetImageDimensions(firstSlide)
	if width%2 != 0 {
		width--
	}
	if height%2 != 0 {
		height--
	}
	if err != nil {
		logger.Error("Failed to get image dimensions", "error", err)
		os.Exit(1)
	}
	cacheDir := filepath.Join(dataDir, "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
			logger.Error("Failed to create cache directory", "error", err)
			os.Exit(1)
		}
	}

	for audioLang, audioList := range audiosLangToPath {
		langDir := filepath.Join(cacheDir, audioLang)
		if _, err := os.Stat(langDir); os.IsNotExist(err) {
			os.Mkdir(langDir, os.ModePerm)
		}
		videoDir := filepath.Join(langDir, "videos")
		if _, err := os.Stat(videoDir); os.IsNotExist(err) {
			if err := os.Mkdir(videoDir, os.ModePerm); err != nil {
				logger.Error("Failed to create video directory", "error", err, "dir", videoDir)
				os.Exit(1)
			}
		}

		if len(slides) != len(audioList) {
			logger.Error("Slide and audio count mismatch",
				"slides", len(slides),
				"audios", len(audioList),
				"language", audioLang)
			os.Exit(1)
		}

		var videoFiles []string

		for i := 0; i < len(slides); i++ {
			slide := slides[i]
			audio := audioList[i]
			outputVideo := fmt.Sprintf("slide%d_video_%s.mp4", i+1, audioLang)
			outputVideoPath := filepath.Join(videoDir, outputVideo)

			iw, ih, err := GetImageDimensions(slide)
			if err != nil {
				logger.Error("Failed to get image dimensions", "error", err)
				os.Exit(1)
			}

			scaleFilter := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", width, height)
			padFilter := fmt.Sprintf("pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1", width, height)
			filterComplex := fmt.Sprintf("%s,%s", scaleFilter, padFilter)

			var cmd *exec.Cmd
			if width != iw || height != ih {
				cmd = exec.Command("ffmpeg", "-loop", "1", "-i", slide, "-i", audio, "-vf", filterComplex, "-c:v", "libx264", "-tune", "stillimage", "-c:a", "mp3", "-b:a", "192k", "-pix_fmt", "yuv420p", "-shortest", outputVideoPath)
			} else {
				cmd = exec.Command("ffmpeg", "-loop", "1", "-i", slide, "-i", audio, "-c:v", "libx264", "-tune", "stillimage", "-c:a", "mp3", "-b:a", "192k", "-pix_fmt", "yuv420p", "-shortest", outputVideoPath)
			}
			logger.Info("Running ffmpeg", "command", cmd.String())

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				logger.Error("Failed to generate video",
					"slide", slide,
					"audio", audio,
					"error", err,
					"ffmpeg_error", stderr.String())
				os.Exit(1)
			}

			videoFiles = append(videoFiles, outputVideoPath)
		}

		outputDir := filepath.Join(dataDir, "out")
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			os.Mkdir(outputDir, os.ModePerm)
		}
		finalOutput := filepath.Join(outputDir, fmt.Sprintf("output-%s.mp4", audioLang))

		concatCmd := BuildFFmpegConcatCommand(videoFiles, finalOutput)
		logger.Info("Concatenating videos", "command", concatCmd.String())

		fmt.Printf("Concatenating videos into %s...\n", finalOutput)
		if err := concatCmd.Run(); err != nil {
			logger.Error("Failed to concatenate videos", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Video created successfully: %s\n", finalOutput)
	}
}
