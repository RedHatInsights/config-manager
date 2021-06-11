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

func ShutdownHTTPServer(ctx context.Context, name string, srv *http.Server) {
	log.Infof("Shutting down %s server", name)
	if err := srv.Shutdown(ctx); err != nil {
		log.Infof("Error shutting down %s server: %e", name, err)
	}
}

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
