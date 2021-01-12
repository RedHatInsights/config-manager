package utils

import (
	"context"
	"net/http"

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
