package render

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog"
)

// RenderPlain writes a plain text response to the response writer, logging the
// response at a level appropriate to the status code.
func RenderPlain(w http.ResponseWriter, r *http.Request, statusCode int, body string, logger zerolog.Logger) {
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

// RenderJSON writes a JSON object response to the response writer, logging the
// response at a level appropriate to the status code.
func RenderJSON(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}, logger zerolog.Logger) {
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

// RenderNone writes only the status code to the response writer, logging the
// response at a level appropriate to the status code.
func RenderNone(w http.ResponseWriter, r *http.Request, statusCode int, logger zerolog.Logger) {
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

// RenderRaw writes sets a Content-Type header and writes body to the response
// writer unmodified, logging the response at a level appropriate to the status
// code.
func RenderRaw(w http.ResponseWriter, r *http.Request, statusCode int, contentType string, body []byte, logger zerolog.Logger) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch {
	case statusCode >= 400:
		logger.Error().Msg("sent HTTP response")
	case statusCode >= 300:
		logger.Info().Msg("sent HTTP response")
	case statusCode >= 200:
		logger.Debug().Msg("sent HTTP response")
	}
}
