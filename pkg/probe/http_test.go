package probe

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guarandoo/neko/pkg/core"
)

func TestHttpProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	probe, err := NewHttpProbe(HttpProbeOptions{
		Url:          server.URL,
		Method:       "GET",
		MaxRedirects: 0,
		Timeout:      5,
	})
	if err != nil {
		t.Fatalf("unable to initialize http probe: %v", err)
	}

	res, err := probe.Probe()
	if err != nil {
		t.Fatalf("probe failed: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Fatalf("unexpected test count: found %v expecting %v", len(res.Tests), 1)
	}

	test := res.Tests[0]
	if test.Status != core.StatusUp {
		t.Fatal("probe did not report target as up")
	}
}

func TestHttpProbeNonOk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	probe, err := NewHttpProbe(HttpProbeOptions{
		Url:          server.URL,
		Method:       "GET",
		MaxRedirects: 0,
		Timeout:      5,
	})
	if err != nil {
		t.Fatalf("unable to initialize http probe: %v", err)
	}

	res, err := probe.Probe()
	if err != nil {
		t.Fatalf("probe failed: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Fatalf("unexpected test count: found %v expecting %v", len(res.Tests), 1)
	}

	test := res.Tests[0]
	if test.Status != core.StatusDown {
		t.Fatal("probe did not report target as down")
	}
}

func testHttpProbeRedirect(t *testing.T, redirectCount int, redirectLimit int) {
	redirects := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if redirects < redirectCount {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		redirects++
	}))
	defer server.Close()

	probe, err := NewHttpProbe(HttpProbeOptions{
		Url:          server.URL,
		Method:       "GET",
		MaxRedirects: redirectLimit,
		Timeout:      5,
	})
	if err != nil {
		t.Fatalf("unable to initialize http probe: %v", err)
	}

	res, err := probe.Probe()
	if err != nil {
		t.Fatalf("probe failed: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Fatalf("unexpected test count: found %v expecting %v", len(res.Tests), 1)
	}

	test := res.Tests[0]
	if redirectCount > redirectLimit {
		if test.Status != core.StatusDown {
			t.Fatal("probe did not report target as down even after redirect limit was exceeded")
		}
	} else {
		if test.Status != core.StatusUp {
			t.Fatal("probe did not report target as up")
		}
	}
}

func TestHttpProbeRedirectLimit(t *testing.T) {
	testHttpProbeRedirect(t, 1, 0)
	testHttpProbeRedirect(t, 1, 1)
	testHttpProbeRedirect(t, 1, 2)
}

func TestHttpProbeInitialize(t *testing.T) {
	_, err := NewHttpProbe(HttpProbeOptions{
		Url: "proto://127.0.0.1",
	})
	if err == nil {
		t.Fatal("NewHttpProbe did not return an error when the scheme is invalid")
	}
	if !errors.Is(err, ErrInvalidUrlScheme) {
		t.Fatal("NewHttpProbe returned an unexpected error")
	}
}
