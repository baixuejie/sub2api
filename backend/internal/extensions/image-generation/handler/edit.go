package imagegeneration

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	pathpkg "path"
	"strconv"
	"strings"

	imagegenerationservice "github.com/Wei-Shaw/sub2api/internal/extensions/image-generation/service"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

const (
	openAIImageEditsPath   = "/v1/images/edits"
	maxImageEditBodyBytes  = 32 << 20
	maxImageEditPartBytes  = 20 << 20
	maxImageEditFileCount  = 4
	maxImageEditFieldBytes = 64 << 10
	preparedEditContextKey = "ext.image_generation.prepared_edit_request"
)

var allowedEditImageTypes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/webp": {},
}

type editUpload struct {
	Name        string
	FileName    string
	ContentType string
	Data        []byte
}

type editMultipartRequest struct {
	Request       imagegenerationservice.GenerationRequest
	InputFidelity string
	Style         string
	Uploads       []editUpload
	Mask          *editUpload
}

type preparedEditRequest struct {
	UserID      int64
	GroupID     int64
	APIKeyID    int64
	Request     imagegenerationservice.GenerationRequest
	Body        []byte
	ContentType string
}

// PrepareEditAPIKey parses only the browser-facing multipart metadata, selects
// a server-side API key, and replaces the request credentials for the existing
// API-key middleware. The original panel credentials are restored on return.
func (h *Handler) PrepareEditAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := middleware.GetAuthSubjectFromContext(c)
		if !ok || subject.UserID <= 0 {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}
		if h == nil || h.service == nil || c.Request == nil || c.Request.Body == nil {
			response.InternalError(c, "Image editing is unavailable")
			c.Abort()
			return
		}

		body, err := readEditBody(c.Request.Body)
		if err != nil {
			if err == errEditBodyTooLarge {
				response.Error(c, http.StatusRequestEntityTooLarge, "Image edit request is too large")
			} else {
				response.BadRequest(c, "Invalid image edit request")
			}
			c.Abort()
			return
		}

		parsed, err := parseEditMultipartRequest(body, c.GetHeader("Content-Type"))
		if err != nil {
			response.BadRequest(c, err.Error())
			c.Abort()
			return
		}
		prepared, err := h.service.Prepare(c.Request.Context(), subject.UserID, parsed.Request)
		if err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}
		if prepared == nil || prepared.APIKey == nil || prepared.APIKey.ID <= 0 || strings.TrimSpace(prepared.APIKey.Key) == "" {
			response.ErrorFrom(c, imagegenerationservice.ErrImageAPIKeyMissing)
			c.Abort()
			return
		}

		forwardBody, forwardContentType, err := buildEditMultipartRequest(prepared.Request, parsed)
		if err != nil {
			response.BadRequest(c, "Invalid image edit request")
			c.Abort()
			return
		}
		c.Set(preparedEditContextKey, preparedEditRequest{
			UserID:      subject.UserID,
			GroupID:     prepared.Request.GroupID,
			APIKeyID:    prepared.APIKey.ID,
			Request:     prepared.Request,
			Body:        forwardBody,
			ContentType: forwardContentType,
		})

		request := c.Request
		originalBody := request.Body
		originalLength := request.ContentLength
		originalContentType := request.Header.Get("Content-Type")
		originalAuthorization := request.Header.Get("Authorization")
		originalAPIKey := request.Header.Get("x-api-key")
		originalGoogleAPIKey := request.Header.Get("x-goog-api-key")
		request.Body = io.NopCloser(bytes.NewReader(forwardBody))
		request.ContentLength = int64(len(forwardBody))
		request.Header.Set("Content-Type", forwardContentType)
		request.Header.Set("Authorization", "Bearer "+prepared.APIKey.Key)
		request.Header.Del("x-api-key")
		request.Header.Del("x-goog-api-key")
		defer func() {
			request.Body = originalBody
			request.ContentLength = originalLength
			restoreHeader(request.Header, "Content-Type", originalContentType)
			restoreHeader(request.Header, "Authorization", originalAuthorization)
			restoreHeader(request.Header, "x-api-key", originalAPIKey)
			restoreHeader(request.Header, "x-goog-api-key", originalGoogleAPIKey)
		}()

		c.Next()
	}
}

// Edit authenticates the prepared server-side key and delegates multipart
// forwarding, routing, moderation and billing to the core Images gateway.
func (h *Handler) Edit(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h == nil || h.gateway == nil {
		response.InternalError(c, "Image editing is unavailable")
		return
	}
	value, exists := c.Get(preparedEditContextKey)
	prepared, preparedOK := value.(preparedEditRequest)
	if !exists || !preparedOK || prepared.UserID != subject.UserID || prepared.GroupID <= 0 || prepared.APIKeyID <= 0 || len(prepared.Body) == 0 {
		response.Forbidden(c, "Image edit request is not prepared")
		return
	}
	apiKey, apiKeyOK := middleware.GetAPIKeyFromContext(c)
	if !apiKeyOK || apiKey == nil || apiKey.ID != prepared.APIKeyID || apiKey.UserID != subject.UserID || apiKey.GroupID == nil || *apiKey.GroupID != prepared.GroupID || apiKey.Group == nil || !apiKey.Group.IsActive() || !apiKey.Group.AllowImageGeneration {
		response.Forbidden(c, "Image editing authorization is invalid")
		return
	}

	request := c.Request
	originalPath := request.URL.Path
	originalRawPath := request.URL.RawPath
	originalBody := request.Body
	originalLength := request.ContentLength
	originalContentType := request.Header.Get("Content-Type")
	defer func() {
		request.URL.Path = originalPath
		request.URL.RawPath = originalRawPath
		request.Body = originalBody
		request.ContentLength = originalLength
		restoreHeader(request.Header, "Content-Type", originalContentType)
	}()

	request.URL.Path = openAIImageEditsPath
	request.URL.RawPath = ""
	request.Body = io.NopCloser(bytes.NewReader(prepared.Body))
	request.ContentLength = int64(len(prepared.Body))
	request.Header.Set("Content-Type", prepared.ContentType)
	h.gateway.Images(c)
}

var errEditBodyTooLarge = fmt.Errorf("image edit body too large")

func readEditBody(body io.Reader) ([]byte, error) {
	if body == nil {
		return nil, fmt.Errorf("missing request body")
	}
	data, err := io.ReadAll(io.LimitReader(body, maxImageEditBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxImageEditBodyBytes {
		return nil, errEditBodyTooLarge
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("image edit request is empty")
	}
	return data, nil
}

func parseEditMultipartRequest(body []byte, contentType string) (editMultipartRequest, error) {
	mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	if err != nil || !strings.EqualFold(mediaType, "multipart/form-data") {
		return editMultipartRequest{}, fmt.Errorf("multipart/form-data content type is required")
	}
	boundary := strings.TrimSpace(params["boundary"])
	if boundary == "" {
		return editMultipartRequest{}, fmt.Errorf("multipart boundary is required")
	}

	parsed := editMultipartRequest{Request: imagegenerationservice.GenerationRequest{N: 1}}
	seen := make(map[string]bool)
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, nextErr := reader.NextPart()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return editMultipartRequest{}, fmt.Errorf("invalid multipart body")
		}
		name := strings.TrimSpace(part.FormName())
		if name == "" {
			_ = part.Close()
			return editMultipartRequest{}, fmt.Errorf("multipart field name is required")
		}
		data, readErr := io.ReadAll(io.LimitReader(part, maxImageEditPartBytes+1))
		fileName := strings.TrimSpace(part.FileName())
		declaredType := normalizeMediaType(part.Header.Get("Content-Type"))
		_ = part.Close()
		if readErr != nil {
			return editMultipartRequest{}, fmt.Errorf("invalid multipart field %s", name)
		}
		if len(data) > maxImageEditPartBytes {
			return editMultipartRequest{}, fmt.Errorf("image file is too large")
		}

		if fileName != "" {
			if name == "mask" {
				if parsed.Mask != nil {
					return editMultipartRequest{}, fmt.Errorf("only one mask is allowed")
				}
				upload, uploadErr := parseEditUpload(name, fileName, declaredType, data)
				if uploadErr != nil {
					return editMultipartRequest{}, uploadErr
				}
				parsed.Mask = &upload
				continue
			}
			if name != "image" && !strings.HasPrefix(name, "image[") && !strings.HasPrefix(name, "image_") {
				return editMultipartRequest{}, fmt.Errorf("unsupported image field %s", name)
			}
			if len(parsed.Uploads) >= maxImageEditFileCount {
				return editMultipartRequest{}, fmt.Errorf("at most %d images are allowed", maxImageEditFileCount)
			}
			upload, uploadErr := parseEditUpload("image", fileName, declaredType, data)
			if uploadErr != nil {
				return editMultipartRequest{}, uploadErr
			}
			parsed.Uploads = append(parsed.Uploads, upload)
			continue
		}

		if len(data) > maxImageEditFieldBytes {
			return editMultipartRequest{}, fmt.Errorf("multipart field %s is too large", name)
		}
		if seen[name] {
			return editMultipartRequest{}, fmt.Errorf("multipart field %s is duplicated", name)
		}
		seen[name] = true
		value := strings.TrimSpace(string(data))
		switch name {
		case "group_id":
			id, parseErr := strconv.ParseInt(value, 10, 64)
			if parseErr != nil || id <= 0 {
				return editMultipartRequest{}, fmt.Errorf("group_id must be a positive integer")
			}
			parsed.Request.GroupID = id
		case "model":
			parsed.Request.Model = value
		case "prompt":
			parsed.Request.Prompt = value
		case "n":
			n, parseErr := strconv.Atoi(value)
			if parseErr != nil || n <= 0 {
				return editMultipartRequest{}, fmt.Errorf("n must be a positive integer")
			}
			parsed.Request.N = n
		case "size":
			parsed.Request.Size = value
		case "quality":
			parsed.Request.Quality = value
		case "output_format":
			parsed.Request.OutputFormat = value
		case "output_compression":
			n, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				return editMultipartRequest{}, fmt.Errorf("output_compression must be an integer")
			}
			parsed.Request.OutputCompression = &n
		case "background":
			parsed.Request.Background = value
		case "moderation":
			parsed.Request.Moderation = value
		case "input_fidelity":
			if value != "" && !strings.EqualFold(value, "low") && !strings.EqualFold(value, "high") {
				return editMultipartRequest{}, fmt.Errorf("input_fidelity is invalid")
			}
			parsed.InputFidelity = value
		case "style":
			if value != "" && !strings.EqualFold(value, "natural") && !strings.EqualFold(value, "vivid") {
				return editMultipartRequest{}, fmt.Errorf("style is invalid")
			}
			parsed.Style = value
		case "response_format":
			if value != "" && !strings.EqualFold(value, "b64_json") {
				return editMultipartRequest{}, fmt.Errorf("response_format must be b64_json")
			}
		default:
			return editMultipartRequest{}, fmt.Errorf("unsupported multipart field %s", name)
		}
	}

	if parsed.Request.GroupID <= 0 {
		return editMultipartRequest{}, fmt.Errorf("group_id is required")
	}
	if strings.TrimSpace(parsed.Request.Model) == "" {
		return editMultipartRequest{}, fmt.Errorf("model is required")
	}
	if strings.TrimSpace(parsed.Request.Prompt) == "" {
		return editMultipartRequest{}, fmt.Errorf("prompt is required")
	}
	if len(parsed.Uploads) == 0 {
		return editMultipartRequest{}, fmt.Errorf("at least one image is required")
	}
	return parsed, nil
}

func parseEditUpload(name, fileName, declaredType string, data []byte) (editUpload, error) {
	if len(data) == 0 {
		return editUpload{}, fmt.Errorf("image file is empty")
	}
	detectedType := normalizeMediaType(http.DetectContentType(data))
	if _, ok := allowedEditImageTypes[detectedType]; !ok {
		return editUpload{}, fmt.Errorf("unsupported image file type")
	}
	if declaredType != "" {
		if _, ok := allowedEditImageTypes[declaredType]; !ok || !sameImageType(declaredType, detectedType) {
			return editUpload{}, fmt.Errorf("image content type does not match its data")
		}
	}
	cleanName := cleanUploadFileName(fileName, detectedType)
	return editUpload{Name: name, FileName: cleanName, ContentType: detectedType, Data: data}, nil
}

func buildEditMultipartRequest(req imagegenerationservice.GenerationRequest, parsed editMultipartRequest) ([]byte, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writeField := func(name, value string) error {
		if strings.TrimSpace(value) == "" {
			return nil
		}
		return writer.WriteField(name, value)
	}
	fields := []struct{ name, value string }{
		{"model", req.Model},
		{"prompt", req.Prompt},
		{"n", strconv.Itoa(req.N)},
		{"size", req.Size},
		{"quality", req.Quality},
		{"output_format", req.OutputFormat},
		{"background", req.Background},
		{"moderation", req.Moderation},
		{"response_format", "b64_json"},
		{"input_fidelity", parsed.InputFidelity},
		{"style", parsed.Style},
	}
	for _, field := range fields {
		if err := writeField(field.name, field.value); err != nil {
			return nil, "", err
		}
	}
	if req.OutputCompression != nil {
		if err := writeField("output_compression", strconv.Itoa(*req.OutputCompression)); err != nil {
			return nil, "", err
		}
	}
	for _, upload := range parsed.Uploads {
		if err := writeEditPart(writer, "image", upload); err != nil {
			return nil, "", err
		}
	}
	if parsed.Mask != nil {
		if err := writeEditPart(writer, "mask", *parsed.Mask); err != nil {
			return nil, "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}

func writeEditPart(writer *multipart.Writer, name string, upload editUpload) error {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     name,
		"filename": upload.FileName,
	}))
	header.Set("Content-Type", upload.ContentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = part.Write(upload.Data)
	return err
}

func normalizeMediaType(value string) string {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return strings.ToLower(strings.TrimSpace(strings.SplitN(value, ";", 2)[0]))
	}
	return strings.ToLower(strings.TrimSpace(mediaType))
}

func sameImageType(left, right string) bool {
	left = normalizeMediaType(left)
	right = normalizeMediaType(right)
	return left == right || (left == "image/jpg" && right == "image/jpeg")
}

func cleanUploadFileName(value, contentType string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = pathpkg.Base(value)
	value = strings.TrimSpace(value)
	if value == "" || value == "." || strings.ContainsAny(value, "\r\n\x00") {
		value = "image"
	}
	if !strings.Contains(value, ".") {
		ext := ".png"
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/webp":
			ext = ".webp"
		}
		value += ext
	}
	return value
}
