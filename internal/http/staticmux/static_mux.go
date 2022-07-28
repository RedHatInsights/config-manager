package staticmux

import (
	"log"
	"net/http"
)

type response struct {
	code int
	body []byte
}

type StaticMux struct {
	handlers map[string]response
}

func (s *StaticMux) AddResponse(path string, responseCode int, responseBody []byte) {
	if s.handlers == nil {
		s.handlers = make(map[string]response)
	}

	s.handlers[path] = response{responseCode, responseBody}
}

func (s *StaticMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, ok := s.handlers[r.URL.Path]
	if ok {
		w.WriteHeader(res.code)
		if _, err := w.Write(res.body); err != nil {
			log.Fatal(err)
		}
		return
	}

	w.WriteHeader(http.StatusNotAcceptable)
}
