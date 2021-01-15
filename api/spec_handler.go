package api

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type ApiSpecServer struct {
	Router       *mux.Router
	SpecFileName string
}

func (s *ApiSpecServer) Routes() {
	s.Router.HandleFunc("/openapi.json", s.handleApiSpec()).Methods(http.MethodGet)
}

func (s *ApiSpecServer) handleApiSpec() http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		file, err := ioutil.ReadFile(s.SpecFileName)
		if err != nil {
			log.Printf("Unable to read API spec file (%s): %s", s.SpecFileName, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(file)
	}
}
