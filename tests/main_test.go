package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var realBaseURL string
var httpClient = &http.Client{Timeout: 5 * time.Second}

func baseURL() string {
	return realBaseURL
}

type successEnvelope struct {
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Meta    map[string]any  `json:"meta"`
}

type errorEnvelope struct {
	Message string            `json:"message"`
	Error   map[string]string `json:"error"`
}

func TestMain(m *testing.M) {
	realBaseURL = strings.TrimSpace(os.Getenv("GOBITE_REAL_BASE_URL"))
	if realBaseURL == "" {
		realBaseURL = "http://localhost:8080"
	}

	healthURL := strings.TrimRight(realBaseURL, "/")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "real tests require a running server (make run). failed to reach %s: %v\n", healthURL, err)
		os.Exit(1)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= http.StatusInternalServerError {
		fmt.Fprintf(os.Stderr, "real tests require a healthy server. %s returned %s\n", healthURL, resp.Status)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func doJSON(t *testing.T, method, path string, payload any, token string) (int, []byte) {
	t.Helper()

	var body io.Reader
	if payload != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			t.Fatalf("encode json: %v", err)
		}
		body = buf
	}

	req, err := http.NewRequest(method, strings.TrimRight(baseURL(), "/")+path, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return resp.StatusCode, respBody
}

func doMultipart(t *testing.T, path, fieldName, filename string, content []byte, token string) (int, []byte) {
	t.Helper()

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile(fieldName, filepath.Base(filename))
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, strings.TrimRight(baseURL(), "/")+path, buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return resp.StatusCode, respBody
}

func decodeSuccess(t *testing.T, body []byte, out any) successEnvelope {
	t.Helper()

	var env successEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			t.Fatalf("decode success data: %v", err)
		}
	}

	return env
}

func decodeError(t *testing.T, body []byte) errorEnvelope {
	t.Helper()

	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}

	return env
}
