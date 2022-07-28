package v1

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog"
)

func renderPlain(w http.ResponseWriter, r *http.Request, statusCode int, body string, logger zerolog.Logger) {
	render.Status(r, statusCode)
	render.PlainText(w, r, body)

	logger = logger.With().Int("status_code", statusCode).Str("status_text", http.StatusText(statusCode)).Str("body", body).Logger()

	switch {
	case statusCode >= 400:
		logger.Error().Msg("sent HTTP response")
	case statusCode >= 300:
		logger.Info().Msg("sent HTTP response")
	case statusCode >= 200:
		logger.Debug().Msg("sent HTTP response")
	}
}

func renderJSON(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}, logger zerolog.Logger) {
	render.Status(r, statusCode)
	render.JSON(w, r, body)

	logger = logger.With().Int("status_code", statusCode).Str("status_text", http.StatusText(statusCode)).Interface("body", body).Logger()

	switch {
	case statusCode >= 400:
		logger.Error().Msg("sent HTTP response")
	case statusCode >= 300:
		logger.Info().Msg("sent HTTP response")
	case statusCode >= 200:
		logger.Debug().Msg("sent HTTP response")
	}
}

func renderNone(w http.ResponseWriter, r *http.Request, statusCode int, logger zerolog.Logger) {
	render.Status(r, statusCode)

	logger = logger.With().Int("status_code", statusCode).Str("status_text", http.StatusText(statusCode)).Logger()

	switch {
	case statusCode >= 400:
		logger.Error().Msg("sent HTTP response")
	case statusCode >= 300:
		logger.Info().Msg("sent HTTP response")
	case statusCode >= 200:
		logger.Debug().Msg("sent HTTP response")
	}
}
