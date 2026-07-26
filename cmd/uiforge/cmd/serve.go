package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/plexusone/uiforge/viewer"
)

var serveCmd = &cobra.Command{
	Use:   "serve [directory]",
	Short: "Start a local development server",
	Long: `Start a local development server with the Dashforge viewer.

The server serves:
  - The embedded Dashforge viewer at /viewer/
  - Your dashboard files from the specified directory

Example:
  uiforge serve ./dashboards
  uiforge serve --port 3000 ./my-project
  uiforge serve --dashboard my-dashboard.json ./data`,
	Args: cobra.MaximumNArgs(1),
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 8080, "Port to serve on")
	serveCmd.Flags().StringP("dashboard", "d", "", "Default dashboard file to load")
	serveCmd.Flags().Bool("no-open", false, "Don't open browser automatically")
}

func runServe(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetInt("port")
	defaultDashboard, _ := cmd.Flags().GetString("dashboard")
	noOpen, _ := cmd.Flags().GetBool("no-open")

	// Determine directory to serve
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving directory: %w", err)
	}

	// Create server mux
	mux := http.NewServeMux()

	// Serve embedded viewer at /viewer/
	mux.Handle("/viewer/", http.StripPrefix("/viewer/", http.FileServer(http.FS(viewer.FS()))))

	// Serve user's directory with CORS headers for development
	mux.Handle("/", &devFileServer{
		root:    absDir,
		handler: http.FileServer(http.Dir(absDir)),
	})

	// SSE endpoint for live reload (future)
	mux.HandleFunc("/api/events", handleSSE)

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost%s", addr)

	if defaultDashboard != "" {
		url = fmt.Sprintf("%s/viewer/?dashboard=../%s", url, defaultDashboard)
	} else {
		url = fmt.Sprintf("%s/viewer/", url)
	}

	fmt.Printf("Dashforge server starting...\n")
	fmt.Printf("  Serving: %s\n", absDir)
	fmt.Printf("  Viewer:  http://localhost:%d/viewer/\n", port)
	if defaultDashboard != "" {
		fmt.Printf("  Dashboard: %s\n", defaultDashboard)
	}
	fmt.Printf("  URL:     %s\n", url)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	// Open browser
	if !noOpen {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(url)
		}()
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return server.ListenAndServe()
}

// devFileServer adds development-friendly headers
type devFileServer struct {
	root    string
	handler http.Handler
}

func (d *devFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for local development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Disable caching for development
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Set content type for JSON files
	if strings.HasSuffix(r.URL.Path, ".json") {
		w.Header().Set("Content-Type", "application/json")
	}

	d.handler.ServeHTTP(w, r)
}

// handleSSE provides Server-Sent Events for live reload
func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Send heartbeat every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Send initial connected message
	if _, err := fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n"); err != nil {
		return
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, "event: heartbeat\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// openBrowser opens the default browser to the given URL
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		log.Printf("Could not detect browser. Please open: %s", url)
		return
	}

	if err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}
