package probe

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guarandoo/neko/pkg/core"
)

func TestHttpProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
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
		t.Fatal("unexpected test count")
	}
}

func TestHttpProbeRedirect(t *testing.T) {
	redirects := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if redirects < 1 {
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

	test := res.Tests[0]
	if test.Status != core.StatusDown {
		t.Fatal("probe did not report down even after redirect limit was exceeded")
	}
}
