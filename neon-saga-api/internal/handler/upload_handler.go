package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	UploadVideoDir = "./uploads/videos"
	UploadImageDir = "./uploads/images"
	MaxVideoSize   = 5 << 30  // 5 GB
	MaxImageSize   = 10 << 20 // 10 MB
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	_ = os.MkdirAll(UploadVideoDir, 0o755)
	_ = os.MkdirAll(UploadImageDir, 0o755)
	return &UploadHandler{}
}

// POST /api/v1/admin/upload/video
func (h *UploadHandler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	h.handleUpload(w, r, "file", UploadVideoDir, MaxVideoSize,
		[]string{".mp4", ".webm", ".mkv", ".mov", ".m4v"}, "video")
}

// POST /api/v1/admin/upload/image
func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	h.handleUpload(w, r, "file", UploadImageDir, MaxImageSize,
		[]string{".jpg", ".jpeg", ".png", ".webp", ".avif"}, "image")
}

func (h *UploadHandler) handleUpload(
	w http.ResponseWriter,
	r *http.Request,
	field, dstDir string,
	maxSize int64,
	allowedExts []string,
	kind string,
) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "файл слишком большой")
		return
	}

	file, header, err := r.FormFile(field)
	if err != nil {
		writeError(w, http.StatusBadRequest, "поле '"+field+"' не найдено")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !contains(allowedExts, ext) {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("недопустимое расширение (разрешены: %s)", strings.Join(allowedExts, ", ")))
		return
	}

	newName := fmt.Sprintf("%s-%d%s", uuid.New().String(), time.Now().Unix(), ext)
	dstPath := filepath.Join(dstDir, newName)

	dst, err := os.Create(dstPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка сохранения файла")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		_ = os.Remove(dstPath)
		writeError(w, http.StatusInternalServerError, "ошибка записи файла")
		return
	}

	urlPath := fmt.Sprintf("/static/%ss/%s", kind, newName)

	writeJSON(w, http.StatusCreated, map[string]any{
		"url":      urlPath,
		"filename": header.Filename,
		"size":     written,
		"kind":     kind,
	})
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}
