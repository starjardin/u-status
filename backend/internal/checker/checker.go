package checker

import (
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
	// Don't follow redirects automatically — treat 3xx as up
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type CheckResult struct {
	StatusCode     *int
	ResponseTimeMs *int
	IsUp           bool
	Error          *string
}

func CheckURL(url string) CheckResult {
	start := time.Now()
	resp, err := httpClient.Get(url)
	elapsed := int(time.Since(start).Milliseconds())

	if err != nil {
		errStr := classifyError(err)
		return CheckResult{IsUp: false, Error: &errStr}
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	isUp := code >= 200 && code < 400

	var errStr *string
	if !isUp {
		s := fmt.Sprintf("HTTP %d", code)
		errStr = &s
	}

	return CheckResult{
		StatusCode:     &code,
		ResponseTimeMs: &elapsed,
		IsUp:           isUp,
		Error:          errStr,
	}
}

func classifyError(err error) string {
	msg := err.Error()
	switch {
	case contains(msg, "no such host"), contains(msg, "DNS"):
		return "DNS resolution failed"
	case contains(msg, "connection refused"):
		return "Connection refused"
	case contains(msg, "timeout"), contains(msg, "deadline exceeded"):
		return "Request timed out"
	case contains(msg, "certificate"), contains(msg, "tls"):
		return "TLS/certificate error"
	default:
		return fmt.Sprintf("Request failed: %s", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsRune(s, substr))
}

func containsRune(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
