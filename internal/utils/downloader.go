package utils

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DetectRangeAndSize checks if `url` supports Range requests and tries to determine total size.
// - ctx: context (for timeout / cancellation).
// - client: an http.Client (if nil, http.DefaultClient with a timeout is used).
// Returns (supportsRange, totalSizeBytes or -1 if unknown, err).
func DetectRangeAndSize(ctx context.Context, client *http.Client, url string) (bool, int64, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	// 1) Try a small ranged GET (bytes=0-0)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, -1, err
	}
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return false, -1, err
	}
	// Always close body
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusPartialContent: // 206 -> Range support
		// Parse Content-Range: "bytes 0-0/12345"
		if cr := resp.Header.Get("Content-Range"); cr != "" {
			if total, ok := parseTotalFromContentRange(cr); ok {
				return true, total, nil
			}
			// malformed Content-Range: fall through to HEAD fallback
		}
		// 206 but missing/invalid Content-Range -> try HEAD to get Content-Length
		if total, ok := tryHeadForSize(ctx, client, url); ok {
			return true, total, nil
		}
		// 206 but no total known (broken server)
		return true, -1, nil

	case http.StatusRequestedRangeNotSatisfiable: // 416 -> server may include */total
		// Parse Content-Range like: "bytes */12345"
		if cr := resp.Header.Get("Content-Range"); cr != "" {
			if total, ok := parseTotalFromContentRange(cr); ok {
				return true, total, nil
			}
		}
		// If Content-Range missing/invalid, fallback to HEAD
		if total, ok := tryHeadForSize(ctx, client, url); ok {
			return true, total, nil
		}
		// 416 but unknown total
		return true, -1, nil

	case http.StatusOK: // 200 -> server does not honor Range; use Content-Length if present
		if cl := resp.Header.Get("Content-Length"); cl != "" {
			if n, err := strconv.ParseInt(cl, 10, 64); err == nil {
				return false, n, nil
			}
			// malformed Content-Length -> try HEAD
		}
		// Try HEAD as fallback
		if total, ok := tryHeadForSize(ctx, client, url); ok {
			return false, total, nil
		}
		// no size available
		return false, -1, nil

	default:
		// Unexpected status code (redirects handled by client). Try HEAD as a safe fallback.
		if total, ok := tryHeadForSize(ctx, client, url); ok {
			// If HEAD indicates Accept-Ranges (rare), report accordingly.
			// But since GET didn't return 206, treat as not supporting range.
			return false, total, nil
		}
		return false, -1, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// parseTotalFromContentRange parses Content-Range header and returns (total, ok).
// Accepts common forms like:
//   "bytes 0-0/12345"
//   "bytes 0-10/104857600"
//   "bytes */12345"   (used with 416)
func parseTotalFromContentRange(cr string) (int64, bool) {
	// Normalize and trim
	cr = strings.TrimSpace(cr)
	// Accept forms: "<unit> <range>/<total>"
	// e.g. "bytes 0-0/1234" or "bytes */1234"
	parts := strings.Fields(cr)
	if len(parts) < 2 {
		return -1, false
	}
	// parts[0] = unit (expect "bytes"), parts[1] = "0-0/1234" or "*/1234"
	rangeAndTotal := parts[1]
	// If unit and rest were joined due to extra spaces, join the remaining parts
	if len(parts) > 2 {
		rangeAndTotal = strings.Join(parts[1:], " ")
	}

	slashIdx := strings.LastIndex(rangeAndTotal, "/")
	if slashIdx == -1 || slashIdx == len(rangeAndTotal)-1 {
		return -1, false
	}
	totalStr := strings.TrimSpace(rangeAndTotal[slashIdx+1:])
	if totalStr == "*" || totalStr == "" {
		return -1, false
	}
	total, err := strconv.ParseInt(totalStr, 10, 64)
	if err != nil || total < 0 {
		return -1, false
	}
	return total, true
}

// tryHeadForSize performs a HEAD request and returns (total, ok).
// It also checks Accept-Ranges header optionally although HEAD can be disabled.
func tryHeadForSize(ctx context.Context, client *http.Client, url string) (int64, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return -1, false
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, false
	}
	defer resp.Body.Close()

	// If HEAD returned Content-Length parse it
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		if n, err := strconv.ParseInt(cl, 10, 64); err == nil {
			return n, true
		}
	}
	// HEAD didn't give size
	return -1, false
}