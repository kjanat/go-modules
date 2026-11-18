//go:build !wasip1 && !wasip2 && !tinygo

package oci

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/scheme/reg"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/ref"
)

func TestPullWITCancelledDuringRegistryResponse(t *testing.T) {
	const mediaType = "application/vnd.oci.image.manifest.v1+json"
	const digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const manifest = `{"schemaVersion":2,"mediaType":"` + mediaType + `","config":{"mediaType":"application/vnd.oci.empty.v1+json","digest":"` + digest + `","size":2},"layers":[{"mediaType":"application/octet-stream","digest":"` + digest + `","size":16}]}`
	for _, tc := range []struct {
		name       string
		path       string
		status     int
		diagnostic error
		stage      string
	}{
		{"manifest error body", "/v2/test/wit/manifests/v1", 404, errs.ErrNotFound, "failed to get manifest"},
		{"blob error body", "/v2/test/wit/blobs/" + digest, 404, errs.ErrNotFound, "failed to fetch blob"},
		{"truncated blob", "/v2/test/wit/blobs/" + digest, 200, io.ErrUnexpectedEOF, "failed to read blob content"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == tc.path:
					w.Header().Set("Content-Length", "16")
					w.WriteHeader(tc.status)
					io.WriteString(w, "partial")
					w.(http.Flusher).Flush()
					if tc.status == 404 {
						<-r.Context().Done()
					}
				case r.URL.Path == "/v2/test/wit/manifests/v1":
					w.Header().Set("Content-Type", mediaType)
					io.WriteString(w, manifest)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			host := strings.TrimPrefix(server.URL, "https://")
			r, err := ref.New(host + "/test/wit:v1")
			if err != nil {
				t.Fatal(err)
			}
			client := *server.Client()
			base := client.Transport
			client.Transport = transportFunc(func(req *http.Request) (*http.Response, error) {
				resp, err := base.RoundTrip(req)
				if err == nil && req.URL.Path == tc.path {
					resp.Body = &readObserver{resp.Body, func(n int, readErr error) {
						// Cancel after the client read the HTTP diagnostic, or
						// just before regclient tries to recover a short read.
						if (tc.status == 404 && n > 0) || errors.Is(readErr, io.ErrUnexpectedEOF) {
							cancel()
						}
					}}
				}
				return resp, err
			})
			rc := regclient.New(regclient.WithRegOpts(reg.WithHTTPClient(&client)))
			data, err := pullWIT(ctx, rc, r)
			if data != nil || !errors.Is(err, tc.diagnostic) || !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), tc.stage) {
				t.Fatalf("got (%q, %v), want %s with %v and cancellation", data, err, tc.stage, tc.diagnostic)
			}
		})
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (fn transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

type readObserver struct {
	io.ReadCloser
	after func(int, error)
}

func (r *readObserver) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.after(n, err)
	return n, err
}
