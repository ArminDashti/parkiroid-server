package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/gin-gonic/gin"
)

type SoundsHandler struct {
	soundsDir string
}

func NewSoundsHandler(soundsDir string) *SoundsHandler {
	return &SoundsHandler{soundsDir: soundsDir}
}

func (handler *SoundsHandler) ListSounds(context *gin.Context) {
	entries := make([]models.SoundManifestEntry, 0)

	if err := os.MkdirAll(handler.soundsDir, 0o755); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read sounds"})
		return
	}

	files, err := os.ReadDir(handler.soundsDir)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read sounds"})
		return
	}

	baseURL := publicAPIBase(context)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".ogg" && ext != ".wav" && ext != ".mp3" {
			continue
		}

		id := strings.TrimSuffix(name, ext)
		format := strings.TrimPrefix(ext, ".")
		path := filepath.Join(handler.soundsDir, name)
		checksum := fileSHA256(path)
		alertType := inferAlertType(id)

		entries = append(entries, models.SoundManifestEntry{
			ID:        id,
			URL:       baseURL + "/sounds/" + id,
			SHA256:    checksum,
			AlertType: alertType,
			Format:    format,
		})
	}

	context.JSON(http.StatusOK, gin.H{"sounds": entries})
}

func (handler *SoundsHandler) GetSound(context *gin.Context) {
	soundID := context.Param("id")
	if soundID == "" || strings.Contains(soundID, "..") || strings.ContainsAny(soundID, `/\`) {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid sound id"})
		return
	}

	for _, ext := range []string{".ogg", ".wav", ".mp3"} {
		path := filepath.Join(handler.soundsDir, soundID+ext)
		if _, err := os.Stat(path); err == nil {
			context.File(path)
			return
		}
	}

	context.JSON(http.StatusNotFound, models.ErrorResponse{Error: "sound not found"})
}

func fileSHA256(path string) string {
	contents, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}

func inferAlertType(id string) string {
	known := []string{
		"bump", "person", "sound_spike", "vehicle_departed",
		"intrusion", "overspeed", "speed_camera", "generic_warning",
	}
	lower := strings.ToLower(id)
	for _, alertType := range known {
		if strings.Contains(lower, alertType) {
			return alertType
		}
	}
	return "generic_warning"
}

func publicAPIBase(context *gin.Context) string {
	scheme := context.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if context.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := context.Request.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = context.Request.Host
	}
	return scheme + "://" + host + apiBasePath
}
