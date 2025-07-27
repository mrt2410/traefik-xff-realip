package traefik_xff_realip_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	plugin "github.com/mrt2410/traefik-xff-realip"
)

func TestNew(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.ExcludedNets = []string{"127.0.0.1/24"}
	cfg.CleanXFF = true

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	testCases := []struct {
		header        string
		desc          string
		xForwardedFor string
		expectedReal  string
		expectedXFF   string
		cleanXFF      bool
	}{
		{
			header:        "X-Forwarded-For",
			desc:          "don't forward, cleanxff true",
			xForwardedFor: "127.0.0.2",
			expectedReal:  "",
			expectedXFF:   "",
			cleanXFF:      true,
		},
		{
			header:        "X-Forwarded-For",
			desc:          "forward, cleanxff true",
			xForwardedFor: "10.0.0.1",
			expectedReal:  "10.0.0.1",
			expectedXFF:   "10.0.0.1",
			cleanXFF:      true,
		},
		{
			header:        "Cf-Connecting-Ip",
			desc:          "forward, cleanxff true",
			xForwardedFor: "10.0.0.1",
			expectedReal:  "10.0.0.1",
			expectedXFF:   "10.0.0.1",
			cleanXFF:      true,
		},
		{
			header:        "X-Forwarded-For",
			desc:          "forward, cleanxff false",
			xForwardedFor: "10.0.0.1",
			expectedReal:  "10.0.0.1",
			expectedXFF:   "10.0.0.1",
			cleanXFF:      false,
		},
		{
			header:        "X-Forwarded-For",
			desc:          "don't forward, cleanxff false",
			xForwardedFor: "127.0.0.2",
			expectedReal:  "",
			expectedXFF:   "127.0.0.2",
			cleanXFF:      false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			cfg.CleanXFF = test.cleanXFF
			handler, err := plugin.New(ctx, next, cfg, "traefik-real-ip")
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set(test.header, test.xForwardedFor)
			// For Cf-Connecting-Ip test case, set both headers
			if test.header == "Cf-Connecting-Ip" {
				req.Header.Set("X-Forwarded-For", "127.0.0.2") // excluded IP
				req.Header.Set("Cf-Connecting-Ip", test.xForwardedFor)
			}

			handler.ServeHTTP(recorder, req)

			assertHeader(t, req, "X-Real-Ip", test.expectedReal)
			assertHeader(t, req, "X-Forwarded-For", test.expectedXFF)
		})
	}
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}
