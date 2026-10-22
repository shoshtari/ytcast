package pipeline

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ytcast/config"
	"ytcast/downloader"
	"ytcast/feed"
	"ytcast/youtube"
)

type Pipeline struct {
	cfg        *config.Config
	downloader *downloader.Downloader
	feedGen    *feed.Generator
	state      *feed.State
	stateMutex sync.RWMutex
}

func New(cfg *config.Config) (*Pipeline, error) {
	dl, err := downloader.New(cfg.Storage.AudioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}

	fg, err := feed.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed generator: %w", err)
	}

	p := &Pipeline{
		cfg:        cfg,
		downloader: dl,
		feedGen:    fg,
		state: &feed.State{
			LastChecked: make(map[string]time.Time),
			Processed:   make(map[string]bool),
		},
	}

	// Load state if exists
	if err := p.loadState(); err != nil {
		log.Printf("Warning: could not load state: %v", err)
	}

	return p, nil
}

func (p *Pipeline) loadState() error {
	statePath := filepath.Join(p.cfg.Storage.FeedDir, "state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()
	return json.Unmarshal(data, p.state)
}

func (p *Pipeline) saveState() error {
	p.stateMutex.RLock()
	defer p.stateMutex.RUnlock()

	data, err := json.MarshalIndent(p.state, "", "  ")
	if err != nil {
		return err
	}

	statePath := filepath.Join(p.cfg.Storage.FeedDir, "state.json")
	return os.WriteFile(statePath, data, 0644)
}

func (p *Pipeline) Run() error {
	log.Println("Starting pipeline run...")

	for _, channelCfg := range p.cfg.Channels {
		if err := p.processChannel(channelCfg); err != nil {
			log.Printf("Error processing channel %s: %v", channelCfg.Name, err)
			continue
		}
	}

	if err := p.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	log.Println("Pipeline run completed")
	return nil
}

func (p *Pipeline) processChannel(channelCfg config.ChannelConfig) error {
	log.Printf("Processing channel: %s (%s)", channelCfg.Name, channelCfg.URL)

	// Get channel info
	channel, err := youtube.GetChannelInfo(channelCfg.URL)
	if err != nil {
		return fmt.Errorf("failed to get channel info: %w", err)
	}
	channel.Title = channelCfg.Name // Use configured name

	// Get latest videos (check last 10)
	videos, err := youtube.GetLatestVideos(channelCfg.URL, 10)
	if err != nil {
		return fmt.Errorf("failed to get latest videos: %w", err)
	}

	log.Printf("Found %d videos for channel %s", len(videos), channel.Title)

	// Process each video
	downloadedIDs := make(map[string]string)
	for _, video := range videos {
		// Check if already processed
		p.stateMutex.RLock()
		processed := p.state.Processed[video.ID]
		p.stateMutex.RUnlock()

		if processed {
			// Still add to the map for feed generation
			if p.downloader.CheckExists(video.ID) {
				downloadedIDs[video.ID] = p.feedGen.GetAudioURL(video.ID)
			}
			continue
		}

		log.Printf("Processing video: %s (%s)", video.Title, video.ID)

		// Download audio
		result, err := p.downloader.DownloadAudio(video.ID, video.URL)
		if err != nil {
			log.Printf("Failed to download video %s: %v", video.ID, err)
			continue
		}

		log.Printf("Downloaded: %s (%d bytes)", result.FilePath, result.FileSize)

		// Mark as processed
		p.stateMutex.Lock()
		p.state.Processed[video.ID] = true
		p.state.LastChecked[channel.ID] = time.Now()
		p.stateMutex.Unlock()

		downloadedIDs[video.ID] = p.feedGen.GetAudioURL(video.ID)
	}

	// Generate RSS feed
	feedPath, err := p.feedGen.GenerateFeed(channel, videos, downloadedIDs)
	if err != nil {
		return fmt.Errorf("failed to generate feed: %w", err)
	}

	log.Printf("Generated feed: %s", feedPath)
	return nil
}

// StartScheduler begins the interval-based polling
func (p *Pipeline) StartScheduler() error {
	interval, err := p.cfg.Polling.Duration()
	if err != nil {
		return fmt.Errorf("invalid polling interval: %w", err)
	}

	// Run immediately first
	if err := p.Run(); err != nil {
		log.Printf("Initial run failed: %v", err)
	}

	// Start ticker
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Scheduler started, checking every %s", interval)

	for range ticker.C {
		if err := p.Run(); err != nil {
			log.Printf("Scheduled run failed: %v", err)
		}
	}

	return nil
}
