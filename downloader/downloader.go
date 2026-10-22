package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Downloader struct {
	outputDir string
}

func New(outputDir string) (*Downloader, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	return &Downloader{outputDir: outputDir}, nil
}

type DownloadResult struct {
	FilePath string
	FileSize int64
}

// DownloadAudio downloads audio from a YouTube video as MP3
func (d *Downloader) DownloadAudio(videoID, videoURL string) (*DownloadResult, error) {
	outputTemplate := filepath.Join(d.outputDir, videoID+".%(ext)s")

	cmd := exec.Command("yt-dlp",
		"-x",                    // Extract audio
		"--audio-format", "mp3", // Convert to MP3
		"--audio-quality", "0", // Best quality
		"-o", outputTemplate,
		"--no-playlist",
		"--no-mtime",
		videoURL,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp download error: %w\nOutput: %s", err, string(output))
	}

	// Find the downloaded file
	expectedPath := filepath.Join(d.outputDir, videoID+".mp3")
	info, err := os.Stat(expectedPath)
	if err != nil {
		return nil, fmt.Errorf("downloaded file not found at %s: %w", expectedPath, err)
	}

	return &DownloadResult{
		FilePath: expectedPath,
		FileSize: info.Size(),
	}, nil
}

// CheckExists checks if an audio file already exists
func (d *Downloader) CheckExists(videoID string) bool {
	path := filepath.Join(d.outputDir, videoID+".mp3")
	_, err := os.Stat(path)
	return err == nil
}

// GetFilePath returns the expected file path for a video
func (d *Downloader) GetFilePath(videoID string) string {
	return filepath.Join(d.outputDir, videoID+".mp3")
}
