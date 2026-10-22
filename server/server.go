package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ytcast/config"
)

type Server struct {
	cfg    *config.Config
	server *http.Server
}

func New(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve audio files
	audioFS := http.FileServer(http.Dir(s.cfg.Storage.AudioDir))
	mux.Handle("/audio/", http.StripPrefix("/audio/", audioFS))

	// Serve RSS feeds
	feedFS := http.FileServer(http.Dir(s.cfg.Storage.FeedDir))
	mux.Handle("/feed/", http.StripPrefix("/feed/", feedFS))

	// Home page with feed links
	mux.HandleFunc("/", s.homeHandler)

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Starting server on %s", addr)
	return s.server.ListenAndServe()
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// List available feeds
	feedFiles, err := os.ReadDir(s.cfg.Storage.FeedDir)
	if err != nil {
		http.Error(w, "Failed to read feeds directory", http.StatusInternalServerError)
		return
	}

	var feedLinks []string
	for _, file := range feedFiles {
		if strings.HasSuffix(file.Name(), ".xml") {
			feedName := strings.TrimSuffix(file.Name(), ".xml")
			feedURL := fmt.Sprintf("%s/feed/%s", s.cfg.Server.BaseURL, file.Name())
			feedLinks = append(feedLinks, fmt.Sprintf(
				`<li><a href="%s">%s</a> - <a href="%s">RSS Feed</a></li>`,
				feedURL, feedName, feedURL,
			))
		}
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>ytcast - YouTube to Podcast</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        ul { list-style: none; padding: 0; }
        li { margin: 10px 0; padding: 10px; background: #f5f5f5; border-radius: 5px; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>📻 ytcast - YouTube to Podcast</h1>
    <p>Available Podcast Feeds:</p>
    <ul>
        %s
    </ul>
    <h2>How to use:</h2>
    <ol>
        <li>Copy the RSS feed URL</li>
        <li>Add it to your podcast app (Apple Podcasts, Spotify, Castbox, etc.)</li>
        <li>Castbox will automatically index new episodes when you publish them</li>
    </ol>
</body>
</html>
`, strings.Join(feedLinks, "\n"))

	w.Write([]byte(html))
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// ServeFile serves a specific file
func ServeFile(w http.ResponseWriter, r *http.Request, path string) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, cleanPath)
}
