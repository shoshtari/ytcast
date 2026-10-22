# ytcast - YouTube to Podcast Pipeline

A lightweight Go service that monitors YouTube channels, downloads audio via yt-dlp, and publishes as podcast RSS feeds.

## Features

- **YouTube Channel Monitoring**: Periodically checks configured channels for new videos
- **Audio Download**: Extracts audio from videos using yt-dlp (MP3 format)
- **RSS Feed Generation**: Creates iTunes-compatible podcast RSS feeds
- **HTTP Server**: Serves audio files and RSS feeds
- **State Persistence**: Tracks processed videos to avoid re-downloading

## Requirements

- Go 1.21+
- yt-dlp (for downloading YouTube videos)

## Installation

```bash
go build -o ytcast .
```

## Configuration

Edit `config.yaml`:

```yaml
channels:
  - name: "Channel Name"
    url: "https://www.youtube.com/@username"

server:
  host: "0.0.0.0"
  port: 8080
  base_url: "http://localhost:8080"

polling:
  interval: 30m

storage:
  audio_dir: "./audio"
  feed_dir: "./feeds"
  
podcast:
  title: "Your Podcast Name"
  description: "Description"
  author: "Author Name"
  email: "author@email.com"
  language: "en-US"
  category: "Technology"
```

## Usage

### Run with scheduling (continuous mode)

```bash
./ytcast
```

This starts the HTTP server and scheduler. The server will:
- Serve audio files at `http://localhost:8080/audio/<video_id>.mp3`
- Serve RSS feeds at `http://localhost:8080/feed/<channel_id>.xml`

### Run once and exit

```bash
./ytcast --once
```

Useful for cron jobs or one-time syncs.

### Custom config file

```bash
./ytcast --config /path/to/config.yaml
```

## Podcast Distribution

### Castbox

Castbox automatically indexes podcast RSS feeds. To publish:

1. Start ytcast: `./ytcast`
2. Get your RSS feed URL: `http://your-server:8080/feed/<channel_id>.xml`
3. Submit the RSS feed to [Castbox](https://castbox.fm)
4. New episodes will be automatically indexed

Alternatively, Castbox may discover your podcast automatically through:
- Apple Podcasts directory
- Search engines

### Other Platforms

Your RSS feed is compatible with all major podcast platforms:
- Apple Podcasts
- Spotify
- Google Podcasts
- Overcast
- Pocket Casts
- And more...

## Architecture

```
ytcast/
├── config/        # Configuration loading
├── youtube/       # YouTube API via yt-dlp
├── downloader/    # Audio download via yt-dlp
├── feed/          # RSS feed generation
├── pipeline/      # Orchestration and state
├── server/        # HTTP server
└── main.go        # Entry point
```

## Memory Efficiency

Go was chosen for this project because:
- **Low memory footprint**: ~10-15MB base, compared to Python's ~30-50MB
- **Efficient garbage collection**: Optimized for long-running services
- **Single binary**: No runtime dependencies
- **Subprocess isolation**: yt-dlp runs in separate processes, keeping memory usage predictable

The real memory consumer is yt-dlp (can use 500MB-2GB during downloads), but that runs independently regardless of which language orchestrates it.

## License

MIT
