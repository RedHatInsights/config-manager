package v2

import (
	"config-manager/internal"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/dispatch"
	"config-manager/internal/http/render"
	"config-manager/internal/instrumentation"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/rs/zerolog/log"
)

// getProfiles returns a list of profiles as filtered by the limit and offset
// query parameters as well as the org ID of the X-Rh-Identity header.
func getProfiles(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Logger()
	logger = logger.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.GetProfilesError()
		render.RenderPlain(w, r, http.StatusBadRequest, "cannot assert X-Rh-Identity header", logger)
		return
	}
	logger = logger.With().Interface("identity", id).Logger()

	var (
		limit  int
		offset int
	)

	for key, val := range map[string]*int{"limit": &limit, "offset": &offset} {
		if r.URL.Query().Has(key) {
			i, err := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
			if err != nil {
				instrumentation.GetProfilesError()
				render.RenderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("cannot parse '%v': %v", key, err), logger)
				return
			}
			*val = int(i)
		}
	}

	total, err := db.CountProfiles(id.Identity.OrgID)
	if err != nil {
		instrumentation.GetProfilesError()
		render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot count profiles: %v", err), logger)
		return
	}
	logger.Debug().Int("total", total).Msg("found profiles")

	profiles, err := db.GetProfiles(id.Identity.OrgID, "", limit, offset)
	if err != nil {
		instrumentation.GetProfilesError()
		render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get profiles: %v", err), logger)
		return
	}

	response := struct {
		Count   int          `json:"count"`
		Limit   int          `json:"limit"`
		Offset  int          `json:"offset"`
		Total   int          `json:"total"`
		Results []db.Profile `json:"results"`
	}{
		Count:   len(profiles),
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		Results: profiles,
	}

	render.RenderJSON(w, r, http.StatusOK, response, logger)
}

// getProfile returns a single profile identified by the "id" path parameter,
// restricted to the profiles available to the identity defined by the
// X-Rh-Identity header.
func getProfile(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Logger()
	logger = logger.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.GetProfileError()
		render.RenderPlain(w, r, http.StatusBadRequest, "cannot assert X-Rh-Identity header", logger)
		return
	}
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

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.GetProfileError()
		render.RenderPlain(w, r, http.StatusBadRequest, "cannot assert X-Rh-Identity header", logger)
		return
	}
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

	dispatch.Dispatch(newProfile)

	render.RenderJSON(w, r, http.StatusCreated, newProfile, logger)
}

// Constructs and returns a playbook suitable for configuring a host to the
// state of the given profile.
func getPlaybook(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Logger()
	logger = logger.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.GetPlaybookError()
		render.RenderPlain(w, r, http.StatusBadRequest, "cannot assert X-Rh-Identity header", logger)
		return
	}
	logger = logger.With().Interface("identity", id).Logger()

	if !r.URL.Query().Has("profile_id") {
		instrumentation.GetPlaybookError()
		render.RenderPlain(w, r, http.StatusBadRequest, "cannot get profile_id query parameter", logger)
		return
	}

	profileID := r.URL.Query().Get("profile_id")

	profile, err := db.GetProfile(profileID)
	if err != nil {
		instrumentation.GetPlaybookError()
		render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get profile with ID: %v", err), logger)
		return
	}

	playbook, err := internal.GeneratePlaybook(profile.StateConfig())
	if err != nil {
		instrumentation.GetPlaybookError()
		render.RenderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot generate playbook: %v", err), logger)
		return
	}

	render.RenderRaw(w, r, http.StatusOK, "application/x-yaml", []byte(playbook), logger)
}
