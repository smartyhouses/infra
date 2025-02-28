package logger

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap/zapcore"
)

const flushTimeout = 30 * time.Second

type HTTPWriter struct {
	ctx        context.Context
	url        string
	httpClient *http.Client
}

func NewHTTPWriter(ctx context.Context, endpoint string) zapcore.WriteSyncer {
	return &HTTPWriter{
		ctx: ctx,
		url: endpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func NewBufferedHTTPWriter(ctx context.Context, endpoint string) zapcore.WriteSyncer {
	httpWriter := &zapcore.BufferedWriteSyncer{
		WS:            NewHTTPWriter(ctx, endpoint),
		Size:          256 * 1024, // 256 kB
		FlushInterval: 5 * time.Second,
	}
	go func() {
		select {
		case <-ctx.Done():
			if err := httpWriter.Stop(); err != nil {
				fmt.Printf("Error stopping HTTP writer: %v\n", err)
			}
		}
	}()

	return httpWriter
}

// Write sends the logs to the HTTP endpoint.
func (h *HTTPWriter) Write(p []byte) (n int, err error) {
	ctx, cancel := context.WithTimeout(h.ctx, flushTimeout)
	defer cancel()

	// Create HTTP request
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewBuffer(p))
	if err != nil || request == nil {
		return 0, fmt.Errorf("error creating HTTP request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	response, err := h.httpClient.Do(request)
	if err != nil {
		return 0, fmt.Errorf("error sending logs: %w", err)
	}

	err = response.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("error closing response body: %w", err)
	}

	return len(p), nil
}

// Sync is required by zapcore.WriteSyncer.
func (h *HTTPWriter) Sync() error {
	return nil
}
