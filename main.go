package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ytcast/config"
	"ytcast/pipeline"
	"ytcast/server"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	runOnce := flag.Bool("once", false, "Run once and exit (no scheduling)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Loaded configuration from %s", *configPath)

	// Create pipeline
	pipe, err := pipeline.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create pipeline: %v", err)
	}

	// Run once mode
	if *runOnce {
		if err := pipe.Run(); err != nil {
			log.Fatalf("Pipeline run failed: %v", err)
		}
		log.Println("Pipeline completed successfully")
		return
	}

	// Start server in a goroutine
	srv := server.New(cfg)
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Handle shutdown gracefully
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start scheduler in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := pipe.StartScheduler(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown or error
	select {
	case <-stop:
		log.Println("Shutdown signal received")
		if err := srv.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		log.Println("Server stopped")
	case err := <-errChan:
		log.Fatalf("Scheduler error: %v", err)
	}
}
