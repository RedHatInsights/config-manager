package v2

import (
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/http/render"
	"config-manager/internal/instrumentation"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/rs/zerolog/log"
)

// getProfile returns a single profile identified by the "id" path parameter,
// restricted to the profiles available to the identity defined by the
// X-Rh-Identity header.
func getProfile(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Logger()
	logger = logger.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id := identity.GetIdentity(r.Context())
	logger = logger.With().Interface("identity", id).Logger()

	profileID := chi.URLParam(r, "id")
	if profileID == "" {
		instrumentation.GetProfileError()
		render.RenderPlain(w, r, http.StatusBadRequest, "cannot get ID from URL", logger)
		return
	}

	var profile *db.Profile
	if profileID == "current" {
		var err error
		var statemap map[string]string
		if err := json.Unmarshal([]byte(config.DefaultConfig.ServiceConfig), &statemap); err != nil {
			instrumentation.GetProfileError()
			render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot unmarshal statemap: %v", err), logger)
			return
		}
		profile, err = db.GetOrInsertCurrentProfile(id.Identity.OrgID, db.NewProfile(id.Identity.OrgID, id.Identity.AccountNumber, statemap))
		if err != nil {
			instrumentation.GetProfileError()
			render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get current profile for org: %v", err), logger)
			return
		}
	} else {
		var err error
		profile, err = db.GetProfile(profileID)
		if err != nil {
			instrumentation.GetProfileError()
			render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get profile with ID: %v", err), logger)
			return
		}
	}

	render.RenderJSON(w, r, http.StatusOK, profile, logger)
}

// createProfile creates and inserts a profile.
func createProfile(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Logger()
	logger = logger.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id := identity.GetIdentity(r.Context())
	logger = logger.With().Interface("identity", id).Logger()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		instrumentation.CreateProfileError()
		render.RenderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("cannot read request body: %v", err), logger)
		return
	}
	defer r.Body.Close()

	var requestedProfile db.Profile
	if err := json.Unmarshal(data, &requestedProfile); err != nil {
		instrumentation.CreateProfileError()
		render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot unmarshal data: %v", err), logger)
		return
	}

	currentProfile, err := db.GetCurrentProfile(id.Identity.OrgID)
	if err != nil {
		instrumentation.CreateProfileError()
		render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get current profile: %v", err), logger)
		return
	}

	newProfile := db.CopyProfile(*currentProfile)
	newProfile.Active = requestedProfile.Active
	newProfile.Insights = requestedProfile.Insights
	newProfile.Remediations = requestedProfile.Remediations
	newProfile.Compliance = requestedProfile.Compliance

	if newProfile.Equal(*currentProfile) {
		render.RenderJSON(w, r, http.StatusNotModified, currentProfile, logger)
		return
	}

	if err := db.InsertProfile(newProfile); err != nil {
		instrumentation.CreateProfileError()
		render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot insert new profile: %v", err), logger)
		return
	}

	render.RenderJSON(w, r, http.StatusCreated, newProfile, logger)
}
