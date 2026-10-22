package youtube

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Video struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Duration     int       `json:"duration"`
	UploadDate   string    `json:"upload_date"`
	ThumbnailURL string    `json:"thumbnail"`
	URL          string    `json:"webpage_url"`
}

type Channel struct {
	ID    string
	Title string
	URL   string
}

// GetChannelInfo extracts channel information from a YouTube channel URL
func GetChannelInfo(channelURL string) (*Channel, error) {
	cmd := exec.Command("yt-dlp",
		"--dump-json",
		"--playlist-end", "1",
		"--no-download",
		channelURL,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp error: %s", string(exitErr.Stderr))
		}
		return nil, err
	}

	var info struct {
		ChannelID    string `json:"channel_id"`
		Channel      string `json:"channel"`
		ChannelURL   string `json:"channel_url"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	return &Channel{
		ID:    info.ChannelID,
		Title: info.Channel,
		URL:   info.ChannelURL,
	}, nil
}

// GetLatestVideos fetches the latest videos from a channel
func GetLatestVideos(channelURL string, limit int) ([]Video, error) {
	args := []string{
		"--dump-json",
		"--flat-playlist",
		"--playlist-end", fmt.Sprintf("%d", limit),
		"--no-download",
		channelURL,
	}

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp error: %s", string(exitErr.Stderr))
		}
		return nil, err
	}

	var videos []Video
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	
	for decoder.More() {
		var entry struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			Duration     int    `json:"duration"`
			ThumbnailURL string `json:"thumbnail"`
			URL          string `json:"url"`
		}

		if err := decoder.Decode(&entry); err != nil {
			continue
		}

		videos = append(videos, Video{
			ID:           entry.ID,
			Title:        entry.Title,
			Duration:     entry.Duration,
			ThumbnailURL: entry.ThumbnailURL,
			URL:          entry.URL,
		})
	}

	return videos, nil
}

// GetVideoInfo fetches detailed information about a specific video
func GetVideoInfo(videoURL string) (*Video, error) {
	cmd := exec.Command("yt-dlp",
		"--dump-json",
		"--no-download",
		videoURL,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp error: %s", string(exitErr.Stderr))
		}
		return nil, err
	}

	var info struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		Duration     int    `json:"duration"`
		UploadDate   string `json:"upload_date"`
		ThumbnailURL string `json:"thumbnail"`
		URL          string `json:"webpage_url"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	return &Video{
		ID:           info.ID,
		Title:        info.Title,
		Description:  info.Description,
		Duration:     info.Duration,
		UploadDate:   info.UploadDate,
		ThumbnailURL: info.ThumbnailURL,
		URL:          info.URL,
	}, nil
}

// ParseUploadDate parses YouTube's upload_date format (YYYYMMDD)
func (v *Video) ParseUploadDate() (time.Time, error) {
	if v.UploadDate == "" {
		return time.Time{}, fmt.Errorf("no upload date")
	}
	return time.Parse("20060102", v.UploadDate)
}
