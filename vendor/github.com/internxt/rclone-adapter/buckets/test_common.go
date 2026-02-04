package buckets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Common test constants used across multiple test files
const (
	TestMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	// Test bucket IDs
	TestBucket1 = "deadbeefdeadbeefdeadbeefdeadbeef"
	TestBucket2 = "cafebabecafebabecafebabecafebabe"
	TestBucket3 = "1234567890abcdef1234567890abcdef"
	TestBucket4 = "fedcba9876543210fedcba9876543210"
	TestBucket5 = "abcdef1234567890abcdef1234567890"
	TestBucket6 = "0123456789abcdef0123456789abcdef"
	TestBucket7 = "aabbccddaabbccddaabbccddaabbccdd"

	TestToken      = "test-token-123"
	TestBasicAuth  = "Basic test-auth"
	TestFolderUUID = "folder-uuid-123"

	TestFileName      = "test-file.txt"
	TestFileNameNoExt = "test-file"
	TestFileID        = "file-id-123"
	TestFileID2       = "file-id"
	TestFileUUID      = "file-uuid-456"
	TestFileUUID2     = "new-file-uuid"
	TestIndex         = "0123456789abcdef00000123456789abcdef00000000123456789abcdef00000000"

	// Thumbnail test constants
	TestThumbFileUUID = "test-file-uuid"
	TestThumbUUID     = "thumb-uuid"
	TestThumbETag     = "\"thumb-etag\""
	TestThumbFileID   = "thumb-file-id"
	TestThumbPath     = "/upload/thumb"
	TestThumbType     = "png"
)

var TestValidPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x0A,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x02, 0x50, 0x58, 0xEA, 0x00, 0x00, 0x00,
	0x17, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x62, 0xF9, 0xCF, 0x80, 0x0F,
	0x30, 0xE1, 0x95, 0x1D, 0xB1, 0xD2, 0x80, 0x00, 0x00, 0x00, 0xFF, 0xFF,
	0x44, 0xCE, 0x01, 0x16, 0x32, 0xD9, 0xD2, 0x3E, 0x00, 0x00, 0x00, 0x00,
	0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

// MockMultiEndpointServer manages multiple HTTP endpoints for integration testing.
type MockMultiEndpointServer struct {
	startHandler           http.HandlerFunc
	transferHandler        http.HandlerFunc
	finishHandler          http.HandlerFunc
	createMetaHandler      http.HandlerFunc
	multipartStartHandler  http.HandlerFunc
	thumbnailHandler       http.HandlerFunc
	server                 *httptest.Server
}

// NewMockMultiEndpointServer creates a new multi-endpoint mock server for testing
func NewMockMultiEndpointServer() *MockMultiEndpointServer {
	m := &MockMultiEndpointServer{}

	mux := http.NewServeMux()

	// Route handlers based on path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Route based on the path
		switch {
		case strings.Contains(path, "/v2/buckets/") && strings.Contains(path, "/files/start"):
			// StartUpload: POST /network/v2/buckets/{bucket}/files/start
			if m.startHandler != nil {
				m.startHandler(w, r)
			} else if m.multipartStartHandler != nil {
				m.multipartStartHandler(w, r)
			}
		case strings.Contains(path, "/v2/buckets/") && strings.Contains(path, "/files/finish"):
			// FinishUpload: POST /network/v2/buckets/{bucket}/files/finish
			if m.finishHandler != nil {
				m.finishHandler(w, r)
			}
		case strings.Contains(path, "/upload"):
			// Transfer: PUT to storage URL
			if m.transferHandler != nil {
				m.transferHandler(w, r)
			}
		case path == "/drive/files":
			// CreateMetaFile: POST /drive/files
			if m.createMetaHandler != nil {
				m.createMetaHandler(w, r)
			}
		case path == "/drive/files/thumbnail":
			// CreateThumbnail: POST /drive/files/thumbnail
			if m.thumbnailHandler != nil {
				m.thumbnailHandler(w, r)
			}
		default:
			http.NotFound(w, r)
		}
	})

	m.server = httptest.NewServer(mux)
	return m
}

// Close shuts down the mock server
func (m *MockMultiEndpointServer) Close() {
	m.server.Close()
}

// URL returns the base URL of the mock server
func (m *MockMultiEndpointServer) URL() string {
	return m.server.URL
}

// SetupSuccessfulUploadMock configures the mock server for a successful single-part upload
func (m *MockMultiEndpointServer) SetupSuccessfulUploadMock() {
	m.startHandler = func(w http.ResponseWriter, r *http.Request) {
		resp := StartUploadResp{
			Uploads: []UploadPart{
				{
					UUID: "part-uuid-123",
					URL:  m.URL() + "/upload/test-url",
					URLs: []string{m.URL() + "/upload/test-url"},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}

	m.transferHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", "\"test-etag-123\"")
		w.WriteHeader(http.StatusOK)
	}

	m.finishHandler = func(w http.ResponseWriter, r *http.Request) {
		resp := FinishUploadResp{
			ID:      TestFileID,
			Bucket:  TestBucket1,
			Index:   "test-index",
			Created: "2025-01-01T00:00:00Z",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}

	m.createMetaHandler = func(w http.ResponseWriter, r *http.Request) {
		resp := CreateMetaResponse{
			UUID:   TestFileUUID,
			FileID: TestFileID,
			Name:   TestFileName,
			Type:   "txt",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// SetupMultipartUploadMock configures the mock server for a successful multipart upload
func (m *MockMultiEndpointServer) SetupMultipartUploadMock() {
	m.multipartStartHandler = func(w http.ResponseWriter, r *http.Request) {
		// Get number of parts from query parameter
		numParts := 3
		if mp := r.URL.Query().Get("multiparts"); mp != "" {
			fmt.Sscanf(mp, "%d", &numParts)
		}

		// Generate URLs for each part
		urls := make([]string, numParts)
		for i := range urls {
			urls[i] = m.URL() + "/upload/multipart"
		}

		resp := StartUploadResp{
			Uploads: []UploadPart{{
				UUID:     "multipart-uuid",
				UploadId: "multipart-upload-id",
				URLs:     urls,
			}},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}

	m.transferHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", "\"part-etag\"")
		w.WriteHeader(http.StatusOK)
	}

	m.finishHandler = func(w http.ResponseWriter, r *http.Request) {
		resp := FinishUploadResp{ID: TestFileID}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}

	m.createMetaHandler = func(w http.ResponseWriter, r *http.Request) {
		resp := CreateMetaResponse{UUID: TestFileUUID, FileID: TestFileID}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
