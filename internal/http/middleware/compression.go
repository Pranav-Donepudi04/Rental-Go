package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"backend-form/m/internal/logger"

	"go.uber.org/zap"
)

// CompressionMiddleware compresses HTTP responses using gzip
func CompressionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip compression for /metrics endpoint (monitoring tools may not handle it)
		if strings.HasPrefix(r.URL.Path, "/metrics") {
			next(w, r)
			return
		}

		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}

		// Wrap response writer to check content type before compressing
		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			request:        r,
		}

		next(gzw, r)

		// Ensure gzip writer is flushed and closed
		if gzw.gz != nil {
			gzw.gz.Flush()
			gzw.gz.Close()
		}
	}
}

// gzipResponseWriter wraps http.ResponseWriter to write compressed data
type gzipResponseWriter struct {
	http.ResponseWriter
	request       *http.Request
	gz            *gzip.Writer
	compressed    bool
	headerWritten bool
}

func (gzw *gzipResponseWriter) WriteHeader(code int) {
	// Check Content-Type when headers are written (before body)
	if !gzw.headerWritten {
		contentType := gzw.Header().Get("Content-Type")
		// Extract base content type (remove charset, etc.)
		if idx := strings.Index(contentType, ";"); idx >= 0 {
			contentType = strings.TrimSpace(contentType[:idx])
		}
		if isCompressible(contentType) {
			// Initialize gzip writer
			gzw.gz = gzip.NewWriter(gzw.ResponseWriter)
			gzw.compressed = true
			gzw.Header().Set("Content-Encoding", "gzip")
			gzw.Header().Set("Vary", "Accept-Encoding")
		}
		gzw.headerWritten = true
	}
	gzw.ResponseWriter.WriteHeader(code)
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
	// If headers haven't been written yet, check Content-Type now
	if !gzw.headerWritten {
		contentType := gzw.Header().Get("Content-Type")
		// Extract base content type (remove charset, etc.)
		if idx := strings.Index(contentType, ";"); idx >= 0 {
			contentType = strings.TrimSpace(contentType[:idx])
		}
		if isCompressible(contentType) {
			// Initialize gzip writer
			gzw.gz = gzip.NewWriter(gzw.ResponseWriter)
			gzw.compressed = true
			gzw.Header().Set("Content-Encoding", "gzip")
			gzw.Header().Set("Vary", "Accept-Encoding")
		}
		gzw.headerWritten = true
	}

	// Write compressed or uncompressed based on decision
	if gzw.compressed && gzw.gz != nil {
		n, err := gzw.gz.Write(b)
		if err != nil {
			logger.Error("Failed to write compressed response", zap.Error(err))
		}
		return n, err
	}

	// Write uncompressed
	return gzw.ResponseWriter.Write(b)
}

func (gzw *gzipResponseWriter) Close() error {
	if gzw.gz != nil {
		return gzw.gz.Close()
	}
	return nil
}

// isCompressible checks if content type should be compressed
func isCompressible(contentType string) bool {
	// Don't compress already compressed formats
	if strings.Contains(contentType, "gzip") ||
		strings.Contains(contentType, "compress") ||
		strings.Contains(contentType, "deflate") ||
		strings.Contains(contentType, "br") {
		return false
	}

	// Compress text-based and JSON content
	compressibleTypes := []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/xml",
		"application/xhtml+xml",
		"image/svg+xml",
	}

	for _, ct := range compressibleTypes {
		if strings.HasPrefix(contentType, ct) {
			return true
		}
	}

	return false
}
