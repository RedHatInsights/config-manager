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
		headers           map[string]string
	}

	type response struct {
		code int
		body string
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
				code: http.StatusOK,
				body: "OK",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			m := StaticMux{}
			m.AddResponse(test.input.path, test.input.res.code, []byte(test.input.res.body))

			reader := strings.NewReader(test.input.req.body)
			req := httptest.NewRequest(test.input.req.method, test.input.req.url, reader)
			for k, v := range test.input.req.headers {
				req.Header.Add(k, v)
			}
			rr := httptest.NewRecorder()
			m.ServeHTTP(rr, req)
			got := response{rr.Code, rr.Body.String()}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}
