package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
)

const maxLoggedBodyBytes = 64 << 10

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if w.body.Len() < maxLoggedBodyBytes {
		remaining := maxLoggedBodyBytes - w.body.Len()
		w.body.Write(data[:min(len(data), remaining)])
	}
	return w.ResponseWriter.Write(data)
}

func (s *Server) requestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger
		if logger == nil {
			logger = slog.Default()
		}

		if shouldLogRequestBody(r.Method) {
			logRequestBody(logger, r)
		}

		recorder := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		if status >= http.StatusBadRequest {
			logger.Error("endpoint returned error",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", status,
				"response", strings.TrimSpace(recorder.body.String()),
				"remote_addr", r.RemoteAddr,
			)
		}
	})
}

func shouldLogRequestBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func logRequestBody(logger *slog.Logger, r *http.Request) {
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if contentType != "application/json" && contentType != "" {
		logger.Debug("endpoint request payload",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"content_type", contentType,
			"content_length", r.ContentLength,
		)
		return
	}
	if r.Body == nil {
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Debug("failed to read request body for logging",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(data))

	logValue := any(string(data))
	var parsed any
	if err := json.Unmarshal(data, &parsed); err == nil {
		logValue = redactJSON(parsed)
	}

	logger.Debug("endpoint request payload",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery,
		"content_type", firstNonEmpty(r.Header.Get("Content-Type"), "application/json"),
		"content_length", r.ContentLength,
		"body", logValue,
	)
}

func redactJSON(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			if isSensitiveKey(key) {
				redacted[key] = "[REDACTED]"
				continue
			}
			redacted[key] = redactJSON(item)
		}
		return redacted
	case []any:
		redacted := make([]any, 0, len(typed))
		for _, item := range typed {
			redacted = append(redacted, redactJSON(item))
		}
		return redacted
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "password") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "authorization")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
