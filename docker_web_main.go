//go:build dockerweb

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/afkarxyz/SpotiFLAC/backend"
)

type jsonError struct {
	Error string `json:"error"`
}

type webServer struct {
	app     *App
	tmpRoot string
	started time.Time
	version string
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		if err := runHealthcheck(); err != nil {
			log.Fatal(err)
		}
		return
	}

	app := NewApp()
	app.startup(context.Background())
	defer app.shutdown(context.Background())

	tmpRoot := strings.TrimSpace(os.Getenv("SPOTIFLAC_TMP_DIR"))
	if tmpRoot == "" {
		tmpRoot = filepath.Join(os.TempDir(), "spotiflac-web")
	}
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	cleanupStaleDownloadDirs(tmpRoot)

	server := &webServer{app: app, tmpRoot: tmpRoot, started: time.Now(), version: backend.AppVersion}
	mux := http.NewServeMux()
	server.registerAPI(mux)
	mux.Handle("/", http.FileServer(http.Dir("frontend/dist")))

	host := strings.TrimSpace(os.Getenv("SPOTIFLAC_HOST"))
	if host == "" {
		host = "0.0.0.0"
	}
	port := strings.TrimSpace(os.Getenv("SPOTIFLAC_PORT"))
	if port == "" {
		port = "8080"
	}

	addr := host + ":" + port
	log.Printf("SpotiFLAC web listening on http://%s", addr)
	if err := http.ListenAndServe(addr, logRequest(mux)); err != nil {
		log.Fatal(err)
	}
}

func (s *webServer) registerAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/docker/status", s.handleDockerStatus)
	mux.HandleFunc("/api/current-ip", s.handleCurrentIP)
	mux.HandleFunc("/api/defaults", s.handleDefaults)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/fonts", s.handleFonts)
	mux.HandleFunc("/api/metadata", s.handleMetadata)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/search-by-type", s.handleSearchByType)
	mux.HandleFunc("/api/streaming-urls", s.handleStreamingURLs)
	mux.HandleFunc("/api/download", s.handleDownload)
	mux.HandleFunc("/api/download/progress", s.handleDownloadProgress)
	mux.HandleFunc("/api/download/queue", s.handleDownloadQueue)
	mux.HandleFunc("/api/download/queue/add", s.handleAddToDownloadQueue)
	mux.HandleFunc("/api/download/queue/failed", s.handleMarkDownloadFailed)
	mux.HandleFunc("/api/download/queue/cancel", s.handleCancelQueued)
	mux.HandleFunc("/api/download/queue/clear-completed", s.handleClearCompleted)
	mux.HandleFunc("/api/download/queue/clear-all", s.handleClearAll)
	mux.HandleFunc("/api/download/queue/skip", s.handleSkipDownloadItem)
	mux.HandleFunc("/api/download/check-files", s.handleCheckFilesExistence)
	mux.HandleFunc("/api/download/m3u8", s.handleCreateM3U8)
	mux.HandleFunc("/api/history/downloads", s.handleDownloadHistory)
	mux.HandleFunc("/api/history/fetches", s.handleFetchHistory)
	mux.HandleFunc("/api/history/fetches/type", s.handleClearFetchHistoryByType)
	mux.HandleFunc("/api/history/recent-fetches", s.handleRecentFetches)
	mux.HandleFunc("/api/history/downloads/item", s.handleDownloadHistoryItem)
	mux.HandleFunc("/api/history/fetches/item", s.handleFetchHistoryItem)
	mux.HandleFunc("/api/status/check", s.handleAPIStatus)
	mux.HandleFunc("/api/status/custom-tidal", s.handleCustomTidalAPI)
	mux.HandleFunc("/api/status/ffmpeg", s.handleFFmpegInstalled)
	mux.HandleFunc("/api/status/ffmpeg/download", s.handleDownloadFFmpeg)
	mux.HandleFunc("/api/availability", s.handleTrackAvailability)
	mux.HandleFunc("/api/lyrics", s.handleLyrics)
	mux.HandleFunc("/api/cover", s.handleCover)
	mux.HandleFunc("/api/header", s.handleHeader)
	mux.HandleFunc("/api/gallery-image", s.handleGalleryImage)
	mux.HandleFunc("/api/avatar", s.handleAvatar)
	mux.HandleFunc("/api/preview", s.handlePreviewURL)
	mux.HandleFunc("/api/isrc", s.handleTrackISRC)
	mux.HandleFunc("/api/open-url", s.handleOpenURL)
	mux.HandleFunc("/api/unsupported", s.handleUnsupported)
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to write json response: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, err error) {
	if err == nil {
		err = errors.New(http.StatusText(status))
	}
	writeJSON(w, status, jsonError{Error: err.Error()})
}

func cleanupStaleDownloadDirs(tmpRoot string) {
	entries, err := os.ReadDir(tmpRoot)
	if err != nil {
		log.Printf("[startup] temp cleanup skipped: %v", err)
		return
	}

	removed := 0
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "download-") {
			continue
		}
		if err := os.RemoveAll(filepath.Join(tmpRoot, entry.Name())); err != nil {
			log.Printf("[startup] failed to remove stale temp dir %q: %v", entry.Name(), err)
			continue
		}
		removed++
	}
	if removed > 0 {
		log.Printf("[startup] removed %d stale download temp dir(s)", removed)
	}
}

func runHealthcheck() error {
	port := strings.TrimSpace(os.Getenv("SPOTIFLAC_PORT"))
	if port == "" {
		port = "8080"
	}

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/api/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		writeErr(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s required", method))
		return false
	}
	return true
}

func (s *webServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ffmpegInstalled, _ := s.app.CheckFFmpegInstalled()
	queue := s.app.GetDownloadQueue()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":              "ok",
		"time":                time.Now().Format(time.RFC3339),
		"version":             s.version,
		"uptime_seconds":      int64(time.Since(s.started).Seconds()),
		"ffmpeg_installed":    ffmpegInstalled,
		"temp_dir":            s.tmpRoot,
		"docker_web":          true,
		"queued_downloads":    queue.QueuedCount,
		"completed_downloads": queue.CompletedCount,
	})
}

func (s *webServer) handleDockerStatus(w http.ResponseWriter, r *http.Request) {
	ffmpegInstalled, _ := s.app.CheckFFmpegInstalled()
	queue := s.app.GetDownloadQueue()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"app": map[string]interface{}{
			"name":    "SpotiFLAC",
			"version": s.version,
			"mode":    "docker-web",
		},
		"runtime": map[string]interface{}{
			"started_at":      s.started.Format(time.RFC3339),
			"uptime_seconds":  int64(time.Since(s.started).Seconds()),
			"temp_dir":        s.tmpRoot,
			"ffmpeg_installed": ffmpegInstalled,
		},
		"queue": queue,
	})
}

func (s *webServer) handleCurrentIP(w http.ResponseWriter, r *http.Request) {
	value, err := s.app.GetCurrentIPInfo()
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeRawJSON(w, value)
}

func (s *webServer) handleDefaults(w http.ResponseWriter, r *http.Request) {
	defaults := s.app.GetDefaults()
	defaults["downloadPath"] = s.tmpRoot
	writeJSON(w, http.StatusOK, defaults)
}

func (s *webServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := s.app.LoadSettings()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, settings)
	case http.MethodPost:
		var settings map[string]interface{}
		if err := decodeJSON(r, &settings); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		settings["downloadPath"] = s.tmpRoot
		settings["operatingSystem"] = "linux/MacOS"
		if err := s.app.SaveSettings(settings); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (s *webServer) handleFonts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fonts, err := s.app.LoadFonts()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, fonts)
	case http.MethodPost:
		var fonts []map[string]interface{}
		if err := decodeJSON(r, &fonts); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		if err := s.app.SaveFonts(fonts); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (s *webServer) handleMetadata(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req SpotifyMetadataRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("[metadata] fetch url=%q batch=%t timeout=%.0fs", req.URL, req.Batch, req.Timeout)
	value, err := s.app.GetSpotifyMetadata(req)
	if err != nil {
		log.Printf("[metadata] failed url=%q error=%v", req.URL, err)
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	log.Printf("[metadata] completed url=%q bytes=%d", req.URL, len(value))
	writeRawJSON(w, value)
}

func (s *webServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req SpotifySearchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("[search] query=%q limit=%d", req.Query, req.Limit)
	resp, err := s.app.SearchSpotify(req)
	if err != nil {
		log.Printf("[search] failed query=%q error=%v", req.Query, err)
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *webServer) handleSearchByType(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req SpotifySearchByTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("[search-by-type] type=%q query=%q limit=%d offset=%d", req.SearchType, req.Query, req.Limit, req.Offset)
	resp, err := s.app.SearchSpotifyByType(req)
	if err != nil {
		log.Printf("[search-by-type] failed type=%q query=%q error=%v", req.SearchType, req.Query, err)
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *webServer) handleStreamingURLs(w http.ResponseWriter, r *http.Request) {
	trackID := r.URL.Query().Get("spotifyTrackID")
	region := r.URL.Query().Get("region")
	value, err := s.app.GetStreamingURLs(trackID, region)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeRawJSON(w, value)
}

func (s *webServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req DownloadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("[download] start service=%q spotify_id=%q track=%q artist=%q", req.Service, req.SpotifyID, req.TrackName, req.ArtistName)

	jobDir, err := os.MkdirTemp(s.tmpRoot, "download-*")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	defer os.RemoveAll(jobDir)
	req.OutputDir = jobDir

	resp, err := s.app.DownloadTrack(req)
	if err != nil || !resp.Success {
		if err == nil {
			err = errors.New(resp.Error)
		}
		log.Printf("[download] failed service=%q spotify_id=%q error=%v", req.Service, req.SpotifyID, err)
		writeErr(w, http.StatusBadGateway, err)
		return
	}

	filePath := resp.File
	if strings.TrimSpace(filePath) == "" {
		log.Printf("[download] failed service=%q spotify_id=%q error=empty file path", req.Service, req.SpotifyID)
		writeErr(w, http.StatusInternalServerError, errors.New("download completed without a file path"))
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}

	metadata, _ := json.Marshal(resp)
	filename := filepath.Base(filePath)
	w.Header().Set("Content-Type", contentTypeForFile(filename))
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	w.Header().Set("X-SpotiFLAC-Response", string(metadata))
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("failed to stream download: %v", err)
		return
	}
	log.Printf("[download] streamed filename=%q size=%d spotify_id=%q", filename, info.Size(), req.SpotifyID)
}

func contentTypeForFile(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".flac":
		return "audio/flac"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	default:
		if value := mime.TypeByExtension(filepath.Ext(name)); value != "" {
			return value
		}
		return "application/octet-stream"
	}
}

func writeRawJSON(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(value))
}

func (s *webServer) handleDownloadProgress(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.app.GetDownloadProgress())
}

func (s *webServer) handleDownloadQueue(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.app.GetDownloadQueue())
}

func (s *webServer) handleAddToDownloadQueue(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		SpotifyID string `json:"spotifyID"`
		TrackName string `json:"trackName"`
		Artist    string `json:"artistName"`
		Album     string `json:"albumName"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"itemID": s.app.AddToDownloadQueue(req.SpotifyID, req.TrackName, req.Artist, req.Album)})
}

func (s *webServer) handleMarkDownloadFailed(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		ItemID string `json:"itemID"`
		Error  string `json:"errorMsg"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	s.app.MarkDownloadItemFailed(req.ItemID, req.Error)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleCancelQueued(w http.ResponseWriter, r *http.Request) {
	s.app.CancelAllQueuedItems()
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleClearCompleted(w http.ResponseWriter, r *http.Request) {
	s.app.ClearCompletedDownloads()
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleClearAll(w http.ResponseWriter, r *http.Request) {
	s.app.ClearAllDownloads()
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleSkipDownloadItem(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		ItemID   string `json:"itemID"`
		FilePath string `json:"filePath"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	s.app.SkipDownloadItem(req.ItemID, req.FilePath)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleCheckFilesExistence(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		OutputDir string                      `json:"outputDir"`
		RootDir   string                      `json:"rootDir"`
		Tracks    []CheckFileExistenceRequest `json:"tracks"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, s.app.CheckFilesExistence(s.tmpRoot, s.tmpRoot, req.Tracks))
}

func (s *webServer) handleCreateM3U8(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleDownloadHistory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := s.app.GetDownloadHistory()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodDelete:
		if err := s.app.ClearDownloadHistory(); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (s *webServer) handleFetchHistory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := s.app.GetFetchHistory()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item backend.FetchHistoryItem
		if err := decodeJSON(r, &item); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		if err := s.app.AddFetchHistory(item); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case http.MethodDelete:
		if err := s.app.ClearFetchHistory(); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (s *webServer) handleClearFetchHistoryByType(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}
	if err := s.app.ClearFetchHistoryByType(r.URL.Query().Get("type")); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleRecentFetches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		value, err := s.app.GetRecentFetches()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, value)
	case http.MethodPost:
		var req struct {
			Payload string `json:"payload"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		if err := s.app.SaveRecentFetches(req.Payload); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (s *webServer) handleDownloadHistoryItem(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}
	if err := s.app.DeleteDownloadHistoryItem(r.URL.Query().Get("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleFetchHistoryItem(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}
	if err := s.app.DeleteFetchHistoryItem(r.URL.Query().Get("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"online": s.app.CheckAPIStatus(r.URL.Query().Get("type"), r.URL.Query().Get("url"))})
}

func (s *webServer) handleCustomTidalAPI(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"online": s.app.CheckCustomTidalAPI(r.URL.Query().Get("url"))})
}

func (s *webServer) handleFFmpegInstalled(w http.ResponseWriter, r *http.Request) {
	installed, err := s.app.CheckFFmpegInstalled()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"installed": installed})
}

func (s *webServer) handleDownloadFFmpeg(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.app.DownloadFFmpeg())
}

func (s *webServer) handleTrackAvailability(w http.ResponseWriter, r *http.Request) {
	value, err := s.app.CheckTrackAvailability(r.URL.Query().Get("spotifyTrackID"))
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeRawJSON(w, value)
}

func (s *webServer) handleLyrics(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req LyricsDownloadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	req.OutputDir = s.tmpRoot
	resp, err := s.app.DownloadLyrics(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *webServer) handleCover(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req CoverDownloadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	req.OutputDir = s.tmpRoot
	resp, err := s.app.DownloadCover(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *webServer) handleHeader(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req HeaderDownloadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	req.OutputDir = s.tmpRoot
	resp, err := s.app.DownloadHeader(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *webServer) handleGalleryImage(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req GalleryImageDownloadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	resp, err := s.app.DownloadGalleryImage(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *webServer) handleAvatar(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req AvatarDownloadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	resp, err := s.app.DownloadAvatar(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *webServer) handlePreviewURL(w http.ResponseWriter, r *http.Request) {
	value, err := s.app.GetPreviewURL(r.URL.Query().Get("trackID"))
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": value})
}

func (s *webServer) handleTrackISRC(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"isrc": s.app.GetTrackISRC(r.URL.Query().Get("spotifyTrackID"))})
}

func (s *webServer) handleOpenURL(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *webServer) handleUnsupported(w http.ResponseWriter, r *http.Request) {
	writeErr(w, http.StatusNotImplemented, errors.New("this desktop-only feature is not available in the Docker web build"))
}
