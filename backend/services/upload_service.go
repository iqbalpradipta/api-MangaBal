package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"scrapingmanga/backend/config"
)

// UploadedFile holds BalStorage file metadata after upload.
type UploadedFile struct {
	FileID      string
	PreviewURL  string
	DownloadURL string
	ThumbnailURL string
	MimeType    string
	Size        int64
}

// UploadService handles file uploads to BalStorage.
type UploadService interface {
	// EnsureFolder returns folder ID, creating it under parentID if not found.
	EnsureFolder(name string, parentID *string) (string, error)
	// UploadFile uploads reader content with given filename+mime to folderID.
	UploadFile(folderID string, filename string, mime string, content io.Reader) (*UploadedFile, error)
}

type uploadService struct {
	cfg    config.BalStorageConfig
	client *http.Client
	token  string
}

func NewUploadService(cfg config.BalStorageConfig) UploadService {
	return &uploadService{
		cfg:    cfg,
		client: &http.Client{Timeout: 180 * time.Second},
	}
}

func (s *uploadService) baseURL() string {
	return strings.TrimRight(s.cfg.BaseURL, "/")
}

// login fetches a Bearer token and stores it.
func (s *uploadService) login() error {
	body, _ := json.Marshal(map[string]string{
		"email":    s.cfg.Email,
		"password": s.cfg.Password,
	})
	resp, err := s.client.Post(s.baseURL()+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("balstorage login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return fmt.Errorf("balstorage login HTTP %d: %s", resp.StatusCode, b)
	}

	var payload struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("balstorage login decode: %w", err)
	}
	if payload.Data.Token == "" {
		return fmt.Errorf("balstorage login: no token in response")
	}
	s.token = payload.Data.Token
	return nil
}

// authHeader returns Authorization header value, logging in if needed.
func (s *uploadService) authHeader() (string, error) {
	if s.token == "" {
		if err := s.login(); err != nil {
			return "", err
		}
	}
	return "Bearer " + s.token, nil
}

// doGet performs authenticated GET.
func (s *uploadService) doGet(path string) (*http.Response, error) {
	auth, err := s.authHeader()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(http.MethodGet, s.baseURL()+path, nil)
	req.Header.Set("Authorization", auth)
	return s.client.Do(req)
}

// doJSON performs authenticated POST with JSON body.
func (s *uploadService) doJSON(path string, payload any) (*http.Response, error) {
	auth, err := s.authHeader()
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, s.baseURL()+path, bytes.NewReader(body))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	return s.client.Do(req)
}

// listFolders returns folders under parentID (nil = root).
func (s *uploadService) listFolders(parentID *string) ([]map[string]any, error) {
	path := "/folders"
	if parentID != nil {
		path += "?parent_id=" + *parentID
	}
	resp, err := s.doGet(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("balstorage list folders HTTP %d", resp.StatusCode)
	}
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return extractItems(result), nil
}

// EnsureFolder returns existing or newly created folder ID.
func (s *uploadService) EnsureFolder(name string, parentID *string) (string, error) {
	folders, err := s.listFolders(parentID)
	if err != nil {
		return "", err
	}
	for _, f := range folders {
		if f["name"] == name {
			return fmt.Sprint(f["id"]), nil
		}
	}

	body := map[string]any{"name": name}
	if parentID != nil {
		body["parent_id"] = *parentID
	}
	resp, err := s.doJSON("/folders", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 409 = race condition, try listing again
	if resp.StatusCode == 409 {
		folders, err = s.listFolders(parentID)
		if err != nil {
			return "", err
		}
		for _, f := range folders {
			if f["name"] == name {
				return fmt.Sprint(f["id"]), nil
			}
		}
		return "", fmt.Errorf("balstorage folder conflict but not found: %s", name)
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return "", fmt.Errorf("balstorage create folder HTTP %d: %s", resp.StatusCode, b)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	data, _ := result["data"].(map[string]any)
	if data == nil {
		return "", fmt.Errorf("balstorage create folder: unexpected payload")
	}
	return fmt.Sprint(data["id"]), nil
}

// UploadFile uploads a single file to folderID and returns metadata.
func (s *uploadService) UploadFile(folderID string, filename string, mime string, content io.Reader) (*UploadedFile, error) {
	auth, err := s.authHeader()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("files", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, content); err != nil {
		return nil, err
	}
	mw.Close()

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/folders/%s/files", s.baseURL(), folderID), &buf)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("balstorage upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return nil, fmt.Errorf("balstorage upload HTTP %d: %s", resp.StatusCode, b)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	items := extractItems(result)
	if len(items) == 0 {
		return nil, fmt.Errorf("balstorage upload: no file data in response")
	}
	item := items[0]
	fileID := resolveString(item, "id", "file_id", "ID")
	if fileID == "" {
		return nil, fmt.Errorf("balstorage upload: cannot resolve file id")
	}

	base := strings.TrimRight(s.cfg.BaseURL, "/")
	return &UploadedFile{
		FileID:       fileID,
		PreviewURL:   fmt.Sprintf("%s/files/%s/preview", base, fileID),
		DownloadURL:  fmt.Sprintf("%s/files/%s/download", base, fileID),
		ThumbnailURL: fmt.Sprintf("%s/files/%s/thumbnail", base, fileID),
		MimeType:     resolveString(item, "mime_type", "mimeType"),
		Size:         resolveInt64(item, "size"),
	}, nil
}

// extractItems mirrors Python's _extract_items.
func extractItems(payload map[string]any) []map[string]any {
	data := payload["data"]
	if arr, ok := data.([]any); ok {
		return toMapSlice(arr)
	}
	if m, ok := data.(map[string]any); ok {
		if nested, ok := m["data"].([]any); ok {
			return toMapSlice(nested)
		}
		if files, ok := m["files"].([]any); ok {
			return toMapSlice(files)
		}
		if m["id"] != nil {
			return []map[string]any{m}
		}
	}
	return nil
}

func toMapSlice(arr []any) []map[string]any {
	out := make([]map[string]any, 0, len(arr))
	for _, v := range arr {
		if m, ok := v.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func resolveString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func resolveInt64(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	}
	return 0
}
