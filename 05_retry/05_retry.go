package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var errFatal = errors.New("fatal")

func doAttempt(client *http.Client, method, url, key string) (int, string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("%w: %v", errFatal, err)
	}
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, "", err
	}

	return resp.StatusCode, resp.Header.Get("Retry-After"), nil
}

func isSuccess(status int) bool {
	return status >= 200 && status <= 399
}
func isRetryable(status int, err error) bool {
	if err != nil {
		if errors.Is(err, errFatal) {
			return false
		}
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return true
		}
		var oe *net.OpError
		if errors.As(err, &oe) {
			return true
		}

		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return true
		}

		return false
	}

	switch status {
	case 429, 500, 502, 503, 504:
		return true
	}
	return false
}

func methodAllowsRetry(method, key string) bool {
	switch method {
	case "GET", "HEAD", "PUT", "DELETE", "OPTIONS", "TRACE":
		return true
	}
	return key != ""
}

func pauseMs(attempt int, retryAfter string) int {
	if retryAfter != "" {
		sec, err := strconv.Atoi(retryAfter)
		if err == nil && sec >= 0 {
			return sec * 1000
		}
	}

	ceiling := 200
	for i := 1; i < attempt; i++ {
		ceiling *= 2
		if ceiling > 2000 {
			ceiling = 2000
			break
		}
	}
	return rand.Intn(ceiling + 1)
}

func main() {

	if len(os.Args) < 2 || strings.HasPrefix(os.Args[1], "-") {
		fmt.Fprintln(os.Stderr, "usage: retry URL [--method M] [--max-attempts N] [--idempotency-key K]")
		os.Exit(2)
	}

	url := os.Args[1]

	fs := flag.NewFlagSet("retry", flag.ExitOnError)
	method := fs.String("method", "GET", "")
	maxAttempts := fs.Int("max-attempts", 5, "")
	key := fs.String("idempotency-key", "", "")
	fs.Parse(os.Args[2:])
	m := strings.ToUpper(*method)

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	attempt := 1
	for {
		status, retryAfter, err := doAttempt(client, m, url, *key)
		if err != nil {
			fmt.Println("attempt", attempt, "error", err)
		} else {
			fmt.Println("attempt", attempt, "status", status)
		}

		again := isRetryable(status, err) &&
			methodAllowsRetry(m, *key) &&
			attempt < *maxAttempts
		if !again {
			result := "failure"
			exitCode := 1
			if isSuccess(status) {
				result = "success"
				exitCode = 0
			}
			fmt.Println("result", result, "attempts", attempt)
			os.Exit(exitCode)
		}

		attempt++
		ms := pauseMs(attempt, retryAfter)
		fmt.Println("sleep_ms", ms)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}

}
