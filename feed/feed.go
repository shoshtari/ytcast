package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ytcast/config"
	"ytcast/youtube"

	"github.com/gorilla/feeds"
)

type Generator struct {
	cfg       *config.Config
	feedDir   string
	audioDir  string
	baseURL   string
	statePath string
}

type State struct {
	LastChecked map[string]time.Time `json:"last_checked"`
	Processed   map[string]bool      `json:"processed"`
}

func New(cfg *config.Config) (*Generator, error) {
	if err := os.MkdirAll(cfg.Storage.FeedDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create feed directory: %w", err)
	}

	return &Generator{
		cfg:       cfg,
		feedDir:   cfg.Storage.FeedDir,
		audioDir:  cfg.Storage.AudioDir,
		baseURL:   cfg.Server.BaseURL,
		statePath: filepath.Join(cfg.Storage.FeedDir, "state.json"),
	}, nil
}

// GenerateFeed creates an RSS feed for a channel with iTunes podcast tags
func (g *Generator) GenerateFeed(channel *youtube.Channel, videos []youtube.Video, downloadedIDs map[string]string) (string, error) {
	now := time.Now()

	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s - %s", g.cfg.Podcast.Title, channel.Title),
		Link:        &feeds.Link{Href: channel.URL},
		Description: g.cfg.Podcast.Description,
		Author:      &feeds.Author{Name: g.cfg.Podcast.Author, Email: g.cfg.Podcast.Email},
		Created:     now,
	}

	// Add episodes as feed items
	for _, video := range videos {
		audioURL, ok := downloadedIDs[video.ID]
		if !ok {
			continue // Skip videos that haven't been downloaded
		}

		// Parse upload date
		uploadDate, err := video.ParseUploadDate()
		if err != nil {
			uploadDate = now
		}

		item := &feeds.Item{
			Title:       video.Title,
			Link:        &feeds.Link{Href: video.URL},
			Description: video.Description,
			Created:     uploadDate,
			Id:          video.ID,
		}

		// Add enclosure for audio file
		item.Enclosure = &feeds.Enclosure{
			Url:    audioURL,
			Length: "0", // Will be updated when we know the file size
			Type:   "audio/mpeg",
		}

		feed.Items = append(feed.Items, item)
	}

	// Generate RSS XML
	rssXML, err := feed.ToRss()
	if err != nil {
		return "", fmt.Errorf("failed to generate RSS feed: %w", err)
	}

	// Add iTunes podcast namespace and tags
	rssXML = g.addITunesTags(rssXML, channel)

	// Save to file
	feedPath := filepath.Join(g.feedDir, fmt.Sprintf("%s.xml", sanitizeFilename(channel.ID)))
	if err := os.WriteFile(feedPath, []byte(rssXML), 0644); err != nil {
		return "", fmt.Errorf("failed to write feed file: %w", err)
	}

	return feedPath, nil
}

// addITunesTags adds iTunes-specific podcast tags to the RSS feed
func (g *Generator) addITunesTags(rssXML string, channel *youtube.Channel) string {
	// Add iTunes namespace to the RSS element
	rssXML = strings.Replace(rssXML, `<rss version="2.0">`,
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`, 1)

	// Add iTunes channel tags after the description
	itunesChannelTags := fmt.Sprintf(`
    <itunes:author>%s</itunes:author>
    <itunes:summary>%s</itunes:summary>
    <itunes:owner>
      <itunes:name>%s</itunes:name>
      <itunes:email>%s</itunes:email>
    </itunes:owner>
    <itunes:category text="%s"/>
    <itunes:explicit>no</itunes:explicit>
    <itunes:image href="%s/logo.jpg"/>`,
		g.cfg.Podcast.Author,
		g.cfg.Podcast.Description,
		g.cfg.Podcast.Author,
		g.cfg.Podcast.Email,
		g.cfg.Podcast.Category,
		g.baseURL)

	rssXML = strings.Replace(rssXML, fmt.Sprintf(`<description>%s</description>`, g.cfg.Podcast.Description),
		fmt.Sprintf(`<description>%s</description>%s`, g.cfg.Podcast.Description, itunesChannelTags), 1)

	return rssXML
}

// GetAudioURL generates the URL for an audio file
func (g *Generator) GetAudioURL(videoID string) string {
	return fmt.Sprintf("%s/audio/%s.mp3", g.baseURL, videoID)
}

// GetFeedURL returns the URL for a channel's RSS feed
func (g *Generator) GetFeedURL(channelID string) string {
	return fmt.Sprintf("%s/feed/%s.xml", g.baseURL, channelID)
}

// sanitizeFilename removes characters that are invalid in filenames
func sanitizeFilename(s string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(s)
}
