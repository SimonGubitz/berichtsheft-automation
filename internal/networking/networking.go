package networking

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"berichtsheft/internal/cache"
)

func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func HTTPRequestWithCacheAndBasicAuth(ctx context.Context, httpClient *http.Client, header http.Header, body io.Reader, apiCache *cache.FileCache, url, username, password string) (io.Reader, error) {

	l := slog.Default().With("layer", "networking")

	valid, err := apiCache.IsValid(url)
	if err != nil && errors.Is(err, cache.ErrCacheLevelNone) {
		return nil, fmt.Errorf("invalid cache: %w", err)
	}

	if valid {

		l.Debug("cache hit", "url", url)

		// File exists and is valid
		// r, err = readCacheL3(fileName)
		r, err := apiCache.GetReader(url)
		if err != nil {
			return nil, fmt.Errorf("HTTPRequestWithCache: Error reading from cache: %w", err)
		}
		return r, nil

	} else {

		l.Debug("cache miss", "url", url)

		rc, err := doRequestWithBasicAuth(ctx, l, httpClient, header, body, url, username, password)
		if err != nil {
			return nil, fmt.Errorf("HTTPRequestWithCache: Error with request: %w", err)
		}
		defer func() {
			if err := rc.Close(); err != nil {
				l.Error("failed to close request body", "Error", err)
			}
		}()

		l.Debug("done request")

		r, err := apiCache.SetReader(url, rc)
		if err != nil {
			l.Warn("failed ot set cache", "url", url)
			return nil, fmt.Errorf("failed to set cache with reader: %w", err)
		}

		return r, nil
	}
}

func doRequestWithBasicAuth(ctx context.Context, l *slog.Logger, httpClient *http.Client, header http.Header, body io.Reader, url, username, password string) (io.ReadCloser, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", url, body)
	if err != nil {
		l.Warn("failed to create request", "url", url)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = header
	req.SetBasicAuth(username, password)

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Warn("error while making request", "url", url)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				slog.Error("failed to close respose body", "Error", err)
			}
		}()

		l.Warn("unexpected status code", "status code", resp.StatusCode, "url", url)

		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	l.Debug("successful request")

	return resp.Body, nil
}
