package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// StartHTTPServer creates an http.Server using handler as the request handler
// and runs ListenAndServe in a goroutine.
func StartHTTPServer(addr, name string, handler *mux.Router) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Infof("Starting %s server:  %s", name, addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.WithFields(logrus.Fields{"error": err}).Fatalf("%s server error", name)
		}
	}()

	return srv
}

// ShutdownHTTPServer calls srv.Shutdown, logging an error should the shutdown
// fail.
func ShutdownHTTPServer(ctx context.Context, name string, srv *http.Server) {
	log.Infof("Shutting down %s server", name)
	if err := srv.Shutdown(ctx); err != nil {
		log.Infof("Error shutting down %s server: %e", name, err)
	}
}

// FilesIntoMap reads dir and for each file found, reads the contents into a
// map, using the filename as the map key.
func FilesIntoMap(dir, pattern string) map[string][]byte {
	filesMap := make(map[string][]byte)

	files, err := filepath.Glob(dir + pattern)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		info, _ := os.Stat(f)
		content, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		filesMap[info.Name()] = content
	}

	return filesMap
}

// VerifyStatePayload checks whether currentState and payload are deeply equal,
// additionally qualifying the equality treating the value of the "insights" key
// with higher precedence; if the "insights" key equals "disabled", all keys in
// payload must be "disabled" or an error is returned.
func VerifyStatePayload(currentState, payload map[string]string) (bool, error) {
	equal := false
	if reflect.DeepEqual(currentState, payload) {
		equal = true
		return equal, nil
	}

	if payload["insights"] == "disabled" {
		for k, v := range payload {
			if v != "disabled" {
				return equal, fmt.Errorf("Service %s must be disabled if insights is disabled", k)
			}
		}
	}

	return equal, nil
}
