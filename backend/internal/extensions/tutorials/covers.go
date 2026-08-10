package tutorials

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	TutorialCoversDir      = "tutorial-covers"
	TutorialCoverMaxBytes  = 5 << 20
	tutorialCoverBodyLimit = TutorialCoverMaxBytes + 1<<20
	tutorialCoverURLPrefix = "/api/v1/tutorials/covers/"
)

var tutorialCoverNamePattern = regexp.MustCompile(`^[a-f0-9]{32}\.(png|jpg|webp)$`)

// CoverHandler serves administrator cover uploads and public cover reads.
type CoverHandler struct {
	dataDir string
}

func NewCoverHandler(dataDir string) *CoverHandler {
	return &CoverHandler{dataDir: dataDir}
}

func (h *CoverHandler) coversDir() string {
	return filepath.Join(h.dataDir, TutorialCoversDir)
}

// Upload handles POST /api/v1/admin/tutorials/covers.
func (h *CoverHandler) Upload(c *gin.Context) {
	if c.Request.ContentLength > tutorialCoverBodyLimit {
		response.BadRequest(c, "cover file is too large")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, tutorialCoverBodyLimit)
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "cover file is required")
		return
	}
	if header.Size <= 0 || header.Size > TutorialCoverMaxBytes {
		response.BadRequest(c, "cover file must be between 1 byte and 5 MB")
		return
	}
	file, err := header.Open()
	if err != nil {
		response.InternalError(c, "failed to open cover file")
		return
	}
	defer file.Close()
	probe := make([]byte, 512)
	n, readErr := io.ReadFull(file, probe)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		response.BadRequest(c, "failed to read cover file")
		return
	}
	contentType := http.DetectContentType(probe[:n])
	contentType, ok := normalizeTutorialCoverType(contentType)
	if !ok {
		response.BadRequest(c, "cover file must be a PNG, JPEG, or WebP image")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.InternalError(c, "failed to seek cover file")
		return
	}
	if err := os.MkdirAll(h.coversDir(), 0750); err != nil {
		response.InternalError(c, "failed to create cover directory")
		return
	}
	filename, err := randomTutorialCoverName(contentType)
	if err != nil {
		response.InternalError(c, "failed to generate cover filename")
		return
	}
	path := filepath.Join(h.coversDir(), filename)
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		response.InternalError(c, "failed to create cover file")
		return
	}
	written, copyErr := io.Copy(dst, io.LimitReader(file, TutorialCoverMaxBytes+1))
	closeErr := dst.Close()
	if copyErr != nil || closeErr != nil || written > TutorialCoverMaxBytes {
		_ = os.Remove(path)
		if written > TutorialCoverMaxBytes {
			response.BadRequest(c, "cover file is too large")
		} else {
			response.InternalError(c, "failed to save cover file")
		}
		return
	}
	response.Success(c, gin.H{
		"filename":     filename,
		"url":          tutorialCoverURLPrefix + filename,
		"content_type": contentType,
		"size":         written,
	})
}

// Serve handles GET /api/v1/tutorials/covers/:filename.
func (h *CoverHandler) Serve(c *gin.Context) {
	filename := c.Param("filename")
	if !tutorialCoverNamePattern.MatchString(filename) || filepath.Base(filename) != filename {
		c.Status(http.StatusNotFound)
		return
	}
	path := filepath.Join(h.coversDir(), filename)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.Status(http.StatusNotFound)
		} else {
			response.InternalError(c, "failed to read cover file")
		}
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil || stat.IsDir() || stat.Size() <= 0 || stat.Size() > TutorialCoverMaxBytes {
		c.Status(http.StatusNotFound)
		return
	}
	expectedType := mime.TypeByExtension(filepath.Ext(filename))
	probe := make([]byte, 512)
	n, readErr := io.ReadFull(file, probe)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		c.Status(http.StatusNotFound)
		return
	}
	contentType, ok := normalizeTutorialCoverType(http.DetectContentType(probe[:n]))
	if !ok || contentType != expectedType {
		c.Status(http.StatusNotFound)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.InternalError(c, "failed to read cover file")
		return
	}
	c.Header("Cache-Control", "public, max-age=86400, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	c.DataFromReader(http.StatusOK, stat.Size(), contentType, file, nil)
}

func normalizeTutorialCoverType(contentType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "image/png":
		return "image/png", true
	case "image/jpeg":
		return "image/jpeg", true
	case "image/webp":
		return "image/webp", true
	default:
		return "", false
	}
}

func randomTutorialCoverName(contentType string) (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("read random filename bytes: %w", err)
	}
	ext := "png"
	switch contentType {
	case "image/jpeg":
		ext = "jpg"
	case "image/webp":
		ext = "webp"
	}
	return hex.EncodeToString(data[:]) + "." + ext, nil
}
