package staticmux

import (
	"log"
	"net/http"
)

type response struct {
	code    int
	body    []byte
	headers map[string][]string
}

type StaticMux struct {
	handlers map[string]response
}

func (s *StaticMux) AddResponse(path string, responseCode int, responseBody []byte, headers map[string][]string) {
	if s.handlers == nil {
		s.handlers = make(map[string]response)
	}

	s.handlers[path] = response{responseCode, responseBody, headers}
}

func (s *StaticMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, ok := s.handlers[r.URL.Path]
	if ok {
		for k, h := range res.headers {
			for _, v := range h {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(res.code)
		if _, err := w.Write(res.body); err != nil {
			log.Fatal(err)
		}
		return
	}

	w.WriteHeader(http.StatusNotAcceptable)
}
