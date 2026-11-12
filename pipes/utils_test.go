package pipes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pipes "github.com/turbot/pipes-sdk-go"
)

func TestIsNotFound(t *testing.T) {
	cases := map[string]struct {
		r    *http.Response
		want bool
	}{
		"nil": {nil, false},
		"404": {&http.Response{StatusCode: http.StatusNotFound}, true},
		"500": {&http.Response{StatusCode: http.StatusInternalServerError}, false},
		"200": {&http.Response{StatusCode: http.StatusOK}, false},
	}
	for name, tc := range cases {
		if got := isNotFound(tc.r); got != tc.want {
			t.Errorf("%s: isNotFound() = %v, want %v", name, got, tc.want)
		}
	}
}

// testClient returns a client whose server answers every request with status.
func testClient(t *testing.T, status int) *PipesClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprintf(w, `{"status":%d,"title":"test"}`, status)
	}))
	t.Cleanup(srv.Close)
	cfg := pipes.NewConfiguration()
	cfg.Servers = pipes.ServerConfigurations{{URL: srv.URL}}
	return &PipesClient{APIClient: pipes.NewAPIClient(cfg), Config: &Config{}}
}
