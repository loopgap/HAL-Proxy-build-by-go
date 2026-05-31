package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		rl := NewRateLimiter(3, time.Minute)
		defer rl.Stop()

		for i := 0; i < 3; i++ {
			if !rl.Allow("10.0.0.1") {
				t.Fatalf("request %d should be allowed", i+1)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		rl := NewRateLimiter(2, time.Minute)
		defer rl.Stop()

		if !rl.Allow("10.0.0.1") {
			t.Fatal("first request should be allowed")
		}
		if !rl.Allow("10.0.0.1") {
			t.Fatal("second request should be allowed")
		}
		if rl.Allow("10.0.0.1") {
			t.Fatal("third request should be blocked")
		}
	})
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)
	defer rl.Stop()

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("10.0.0.1") {
		t.Fatal("second request should be blocked within window")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("request after window reset should be allowed")
	}
}

func TestRateLimiterMultipleIPIsolation(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	defer rl.Stop()

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first IP first request should be allowed")
	}
	if rl.Allow("10.0.0.1") {
		t.Fatal("first IP second request should be blocked")
	}
	if !rl.Allow("10.0.0.2") {
		t.Fatal("second IP first request should be allowed")
	}
	if rl.Allow("10.0.0.2") {
		t.Fatal("second IP second request should be blocked")
	}
}

func TestRateLimiterWithBurst(t *testing.T) {
	rl := NewRateLimiterWithBurst(2, time.Minute, 1)
	defer rl.Stop()

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}
	if !rl.Allow("10.0.0.1") {
		t.Fatal("second request should be allowed")
	}
	if rl.Allow("10.0.0.1") {
		t.Fatal("third request should be blocked")
	}
}

func TestRateLimiterMaxEntriesPerIP(t *testing.T) {
	rl := NewRateLimiter(maxEntriesPerIP+100, time.Minute)
	defer rl.Stop()

	for i := 0; i < maxEntriesPerIP; i++ {
		if !rl.Allow("10.0.0.1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	if rl.Allow("10.0.0.1") {
		t.Fatal("request beyond maxEntriesPerIP should be blocked")
	}
}

func TestRateLimiterStop(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	done := make(chan struct{})
	go func() {
		rl.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop() did not return")
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")

	rl.mu.RLock()
	count := len(rl.requests["10.0.0.1"])
	rl.mu.RUnlock()
	if count != 2 {
		t.Fatalf("expected 2 entries, got %d", count)
	}

	rl.Stop()
}

func TestRateLimiterAllowFiltersExpired(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)
	defer rl.Stop()

	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")

	if rl.Allow("10.0.0.1") {
		t.Fatal("third request should be blocked within window")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("request after window should be allowed, expired entries filtered")
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	defer rl.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Allow("10.0.0.1")
		}()
	}
	wg.Wait()
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("allows request within limit", func(t *testing.T) {
		rl := NewRateLimiter(5, time.Minute)
		defer rl.Stop()

		handler := RateLimit(rl, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("blocks request over limit", func(t *testing.T) {
		rl := NewRateLimiter(1, time.Minute)
		defer rl.Stop()

		handler := RateLimit(rl, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("first request: expected status %d, got %d", http.StatusOK, rr.Code)
		}

		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rr.Code)
		}
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		rl := NewRateLimiter(1, time.Minute)
		defer rl.Stop()

		handler := RateLimit(rl, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		req1.RemoteAddr = "10.0.0.1:12345"
		rr1 := httptest.NewRecorder()

		handler.ServeHTTP(rr1, req1)
		if rr1.Code != http.StatusOK {
			t.Fatalf("IP1 first request: expected status %d, got %d", http.StatusOK, rr1.Code)
		}

		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.RemoteAddr = "10.0.0.2:12345"
		rr2 := httptest.NewRecorder()

		handler.ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusOK {
			t.Errorf("IP2 first request: expected status %d, got %d", http.StatusOK, rr2.Code)
		}
	})
}
