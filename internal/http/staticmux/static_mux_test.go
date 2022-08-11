package staticmux

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStaticMuxServeHTTP(t *testing.T) {
	type request struct {
		method, url, body string
		headers           map[string][]string
	}

	type response struct {
		code    int
		body    string
		headers map[string][]string
	}

	tests := []struct {
		description string
		input       struct {
			path string
			res  response
			req  request
		}
		want response
	}{
		{
			input: struct {
				path string
				res  response
				req  request
			}{
				path: "/test",
				req: request{
					method: http.MethodGet,
					url:    "/test",
					body:   "",
				},
				res: response{
					code: http.StatusOK,
					body: "OK",
				},
			},
			want: response{
				code:    http.StatusOK,
				body:    "OK",
				headers: map[string][]string{},
			},
		},
		{
			description: "empty JSON object",
			input: struct {
				path string
				res  response
				req  request
			}{
				path: "/test",
				req: request{
					method: http.MethodGet,
					url:    "/test",
					body:   "{}",
				},
				res: response{
					code: http.StatusOK,
					body: "{}",
					headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
				},
			},
			want: response{
				code: http.StatusOK,
				body: "{}",
				headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
			},
		},
		{
			description: "non-empty JSON object",
			input: struct {
				path string
				res  response
				req  request
			}{
				path: "/test",
				req: request{
					method: http.MethodGet,
					url:    "/test",
					body:   "",
				},
				res: response{
					code: http.StatusOK,
					body: "{\"count\":1}",
					headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
				},
			},
			want: response{
				code: http.StatusOK,
				body: "{\"count\":1}",
				headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			m := StaticMux{}
			m.AddResponse(test.input.path, test.input.res.code, []byte(test.input.res.body), test.input.res.headers)

			reader := strings.NewReader(test.input.req.body)
			req := httptest.NewRequest(test.input.req.method, test.input.req.url, reader)
			for k, h := range test.input.req.headers {
				for _, v := range h {
					req.Header.Add(k, v)
				}
			}
			rr := httptest.NewRecorder()
			m.ServeHTTP(rr, req)
			got := response{rr.Code, rr.Body.String(), rr.Header().Clone()}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}
