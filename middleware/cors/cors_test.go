package cors

import (
	"context"
	"net/http"
	"strings"
	"testing"

	config "github.com/go-kratos/gateway/api/gateway/config/v1"
	v1 "github.com/go-kratos/gateway/api/gateway/middleware/cors/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

func buildConfig(origins []string) *config.Middleware {
	v, err := anypb.New(&v1.Cors{
		AllowOrigins: origins,
	})
	if err != nil {
		panic(err)
	}
	return &config.Middleware{Options: v}
}

func TestCors(t *testing.T) {
	tests := []struct {
		Config     *config.Middleware
		Origin     string
		Method     string
		StatusCode int
	}{
		{
			Config:     &config.Middleware{},
			Method:     "POST",
			StatusCode: 200,
		},
		{
			Config:     &config.Middleware{},
			Method:     "OPTIONS",
			StatusCode: 403,
		},
		{
			Config:     buildConfig([]string{".google.com"}),
			Origin:     "https://www.youtube.com",
			Method:     "OPTIONS",
			StatusCode: 403,
		},
		{
			Config:     buildConfig([]string{".google.com"}),
			Origin:     "https://www.google.com",
			Method:     "OPTIONS",
			StatusCode: 200,
		},
	}
	for no, test := range tests {
		m, err := Middleware(test.Config)
		if err != nil {
			t.Fatal(err)
		}
		do := m(func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return newResponse(200, make(http.Header))
		})
		{
			req, err := http.NewRequest(test.Method, "/foo", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set(corsOriginHeader, test.Origin)
			resp, err := do(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != test.StatusCode {
				t.Fatalf("%d want %d but got %d", no, test.StatusCode, resp.StatusCode)
			}
			if resp.StatusCode != 200 {
				continue
			}
			if test.Method == "OPTIONS" {
				// preflightHeaders
				if v := resp.Header.Get(corsVaryHeader); v != corsOriginHeader {
					t.Fatalf("%d want %s but got %s", no, corsOriginHeader, v)
				}
				if v := resp.Header.Get(corsAllowCredentialsHeader); v != "true" {
					t.Fatalf("%d want %s but got %s", no, "true", v)
				}
				if v := resp.Header.Get(corsAllowMethodsHeader); v != strings.Join(defaultCorsMethods, ",") {
					t.Fatalf("%d want %s but got %s", no, defaultCorsMethods, v)
				}
				if v := resp.Header.Get(corsAllowHeadersHeader); v != strings.Join(defaultCorsHeaders, ",") {
					t.Fatalf("%d want %s but got %s", no, defaultCorsHeaders, v)
				}
				if v := resp.Header.Get(corsMaxAgeHeader); v != "600" {
					t.Fatalf("%d want %s but got %s", no, "600", v)
				}
			} else {
				// normalHeaders
				if v := resp.Header.Get(corsVaryHeader); v != corsOriginHeader {
					t.Fatalf("%d want %s but got %s", no, corsOriginHeader, v)
				}
				if v := resp.Header.Get(corsAllowCredentialsHeader); v != "true" {
					t.Fatalf("%d want %s but got %s", no, "true", v)
				}
				if v := resp.Header.Get(corsAllowMethodsHeader); v != strings.Join(defaultCorsMethods, ",") {
					t.Fatalf("%d want %s but got %s", no, defaultCorsMethods, v)
				}
			}
		}
	}
}
