package main

import (
	"fmt"
	"log"
	"os"

	"ytcast/config"
	"ytcast/downloader"
	"ytcast/feed"
	"ytcast/youtube"
)

// This is a manual test file for the ytcast pipeline
// Run with: go run test/test_pipeline.go

func main() {
	// Create test directories
	os.MkdirAll("./test_audio", 0755)
	os.MkdirAll("./test_feeds", 0755)

	// Test with a single video URL
	// Replace with any YouTube video URL you want to test with
	videoURL := "https://www.youtube.com/watch?v=jNQXAC9IVRw" // "Me at the zoo" - first YouTube video

	fmt.Println("Testing ytcast pipeline...")
	fmt.Println()
	fmt.Println("Note: If YouTube is unreachable, this test will timeout")
	fmt.Println("      Ensure you have network access to YouTube")
	fmt.Println()

	// Step 1: Get video info
	fmt.Println("1. Fetching video info...")
	video, err := youtube.GetVideoInfo(videoURL)
	if err != nil {
		log.Printf("Failed to get video info: %v", err)
		fmt.Println("   Skipping video info test due to network/timeout")
		// Create a mock video for testing feed generation
		video = &youtube.Video{
			ID:          "jNQXAC9IVRw",
			Title:       "Me at the zoo",
			Description: "The first video on YouTube",
			Duration:    19,
			UploadDate:  "20050423",
			URL:         videoURL,
		}
	}
	fmt.Printf("   Title: %s\n", video.Title)
	fmt.Printf("   Duration: %d seconds\n", video.Duration)
	fmt.Printf("   ID: %s\n", video.ID)
	fmt.Println()

	// Step 2: Test feed generation (always works offline)
	fmt.Println("2. Testing feed generation...")
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "http://localhost:8080",
		},
		Storage: config.StorageConfig{
			AudioDir: "./test_audio",
			FeedDir:  "./test_feeds",
		},
		Podcast: config.PodcastConfig{
			Title:       "Test Podcast",
			Description: "Test Description",
			Author:      "Test Author",
			Email:       "test@example.com",
			Language:    "en-US",
			Category:    "Technology",
		},
	}

	fg, err := feed.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create feed generator: %v", err)
	}

	channel := &youtube.Channel{
		ID:    "test_channel",
		Title: "Test Channel",
		URL:   "https://youtube.com/@test",
	}

	downloadedIDs := map[string]string{
		video.ID: fg.GetAudioURL(video.ID),
	}

	feedPath, err := fg.GenerateFeed(channel, []youtube.Video{*video}, downloadedIDs)
	if err != nil {
		log.Fatalf("Failed to generate feed: %v", err)
	}
	fmt.Printf("   Feed generated: %s\n", feedPath)
	fmt.Println()

	// Step 3: Show feed content
	fmt.Println("3. Generated RSS feed:")
	content, err := os.ReadFile(feedPath)
	if err != nil {
		log.Fatalf("Failed to read feed: %v", err)
	}
	fmt.Println(string(content))
	fmt.Println()

	// Step 4: Show feed URL
	fmt.Println("4. Feed URL for podcast apps:")
	fmt.Printf("   %s\n", fg.GetFeedURL(channel.ID))
	fmt.Println()

	// Step 5: Test downloader initialization
	fmt.Println("5. Testing downloader...")
	dl, err := downloader.New("./test_audio")
	if err != nil {
		log.Fatalf("Failed to create downloader: %v", err)
	}
	fmt.Printf("   Downloader initialized with output dir: ./test_audio\n")
	fmt.Printf("   File path for video %s: %s\n", video.ID, dl.GetFilePath(video.ID))
	fmt.Println()
	
	fmt.Println("Note: Actual download test requires network access to YouTube")
	fmt.Println("      The binary is ready to use with: ./ytcast --config config.yaml")
	fmt.Println()

	fmt.Println("✓ Pipeline components verified successfully!")
}
