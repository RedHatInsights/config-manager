package v1

import (
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/infrastructure/persistence/inventory"
	"config-manager/internal"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/instrumentation"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/rs/zerolog/log"
)

// accountState represents a row from the deprecated "account_state" table.
type accountState struct {
	Account    string            `json:"account"`
	State      map[string]string `json:"state"`
	ID         string            `json:"id"`
	Label      string            `json:"label"`
	ApplyState bool              `json:"apply_state"`
	OrgID      string            `json:"org_id"`
}

// stateArchive represents a row from the deprecated "state_archives" table.
type stateArchive struct {
	Account   string            `json:"account"`
	ID        string            `json:"id"`
	Label     string            `json:"label"`
	Initiator string            `json:"initiator"`
	CreatedAt time.Time         `json:"created_at"`
	State     map[string]string `json:"state"`
	OrgID     string            `json:"org_id"`
}

// postStates records a new state into the account states table and dispatches
// work to clients if a update is required to bring the client into comformance
// with the given state.
func postStates(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.UpdateAccountStateError()
		renderPlain(w, r, http.StatusBadRequest, "unable to assert X-Rh-Identity header", logger)
		return
	}
	logger = log.With().Interface("identity", id).Logger()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		instrumentation.UpdateAccountStateError()
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("unable to read request body: %v", err), logger)
		return
	}
	defer r.Body.Close()

	var statemap map[string]string
	if err := json.Unmarshal(data, &statemap); err != nil {
		instrumentation.UpdateAccountStateError()
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("unable to unmarshal json: %v", err), logger)
		return
	}

	currentProfile, err := db.GetCurrentProfile(id.Identity.OrgID)
	if err != nil {
		instrumentation.UpdateAccountStateError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to get current profile for account: %v", err), logger)
		return
	}
	logger.Trace().Interface("current_profile", currentProfile).Msg("found current profile")

	equal, err := internal.VerifyStatePayload(currentProfile.StateConfig(), statemap)
	if err != nil {
		instrumentation.UpdateAccountStateError()
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("unable to verify state map: %v", err), logger)
		return
	}

	if equal {
		resp := accountState{
			Account:    currentProfile.AccountID.String,
			ApplyState: currentProfile.Active,
			ID:         currentProfile.ID.String(),
			Label:      currentProfile.Label.String,
			OrgID:      currentProfile.OrgID.String,
			State:      currentProfile.StateConfig(),
		}
		renderJSON(w, r, http.StatusOK, resp, logger)
		return
	}

	newProfile := db.CopyProfile(*currentProfile)
	newProfile.SetStateConfig(statemap)
	newProfile.OrgID.Valid = id.Identity.OrgID != ""
	newProfile.OrgID.String = id.Identity.OrgID
	logger.Trace().Interface("new_profile", newProfile).Msg("created new profile")

	if err := db.InsertProfile(newProfile); err != nil {
		instrumentation.UpdateAccountStateError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to insert new profile: %v", err), logger)
		return
	}

	var clients []internal.Host
	inventoryClient := inventory.NewInventoryClient()
	inventoryResp, err := inventoryClient.GetInventoryClients(r.Context(), 1)
	if err != nil {
		instrumentation.UpdateAccountStateError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("uanble to get inventory clients: %v", err), logger)
		return
	}
	clients = append(clients, inventoryResp.Results...)

	for len(clients) < inventoryResp.Total {
		page := inventoryResp.Page + 1
		res, err := inventoryClient.GetInventoryClients(r.Context(), page)
		if err != nil {
			instrumentation.UpdateAccountStateError()
			renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to get inventory clients: %v", err), logger)
			return
		}
		clients = append(clients, res.Results...)
	}

	go func() {
		internal.ApplyProfile(r.Context(), &newProfile, clients, func(results []dispatcher.RunCreated) {
			logger.Info().Interface("results", results).Msg("response from dispatcher")
		})
		if err != nil {
			instrumentation.UpdateAccountStateError()
			logger.Error().Err(err).Msg("error applying state")
		}
	}()

	resp := accountState{
		Account:    newProfile.AccountID.String,
		ApplyState: newProfile.Active,
		ID:         newProfile.ID.String(),
		Label:      newProfile.Label.String,
		OrgID:      newProfile.OrgID.String,
		State:      newProfile.StateConfig(),
	}
	renderJSON(w, r, http.StatusOK, resp, logger)
}

// getStates returns a list of state archive records as filtered by the sortBy,
// limit and offset query parameters and the account number of the X-Rh-Identity
// header.
func getStates(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Logger()

	logger = logger.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.GetStateChangesError()
		renderPlain(w, r, http.StatusBadRequest, "unable to assert X-Rh-Identity header", logger)
		return
	}
	logger = log.With().Interface("identity", id).Logger()

	var (
		sortBy string
		limit  int
		offset int
	)

	if r.URL.Query().Has("sortBy") {
		sortBy = r.URL.Query().Get("sortBy")
	}

	for key, val := range map[string]*int{"limit": &limit, "offset": &offset} {
		if r.URL.Query().Has(key) {
			i, err := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
			if err != nil {
				instrumentation.GetStateChangesError()
				renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("unable to parse '%v': %v", key, err), logger)
				return
			}
			*val = int(i)
		}
	}

	total, err := db.CountProfiles(id.Identity.OrgID)
	if err != nil {
		instrumentation.GetStateChangesError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to count profiles: %v", err), logger)
		return
	}
	logger.Debug().Int("total", total).Msg("found profiles for account")

	profiles, err := db.GetProfiles(id.Identity.OrgID, sortBy, limit, offset)
	if err != nil {
		instrumentation.GetStateChangesError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to get profiles: %v", err), logger)
		return
	}

	states := make([]stateArchive, 0, len(profiles))

	for _, profile := range profiles {
		s := stateArchive{
			Account:   profile.AccountID.String,
			ID:        profile.ID.String(),
			Label:     profile.Label.String,
			Initiator: profile.Creator.String,
			CreatedAt: profile.CreatedAt.UTC(),
			State:     make(map[string]string),
			OrgID:     profile.OrgID.String,
		}
		s.State = profile.StateConfig()

		states = append(states, s)
	}

	response := struct {
		Count   int            `json:"count"`
		Limit   int            `json:"limit"`
		Offset  int            `json:"offset"`
		Total   int            `json:"total"`
		Results []stateArchive `json:"results"`
	}{
		Count:   len(profiles),
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		Results: states,
	}

	renderJSON(w, r, http.StatusOK, response, logger)
}

// getCurrentState returns the current account state by selecting the current
// profile from the profiles database.
func getCurrentState(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Logger()

	logger = logger.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.GetAccountStateError()
		renderPlain(w, r, http.StatusBadRequest, "unable to assert X-Rh-Identity", logger)
		return
	}
	logger = log.With().Interface("identity", id).Logger()

	var defaultState map[string]string
	if err := json.Unmarshal([]byte(config.DefaultConfig.ServiceConfig), &defaultState); err != nil {
		instrumentation.GetAccountStateError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to unmarshal json: %v", err), logger)
		return
	}

	profile, err := db.GetOrInsertCurrentProfile(id.Identity.OrgID, db.NewProfile(id.Identity.OrgID, id.Identity.AccountNumber, defaultState))
	if err != nil {
		instrumentation.GetAccountStateError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to get profile: %v", err), logger)
		return
	}

	resp := accountState{
		Account:    profile.AccountID.String,
		State:      profile.StateConfig(),
		ID:         profile.ID.String(),
		Label:      profile.Label.String,
		ApplyState: profile.Active,
		OrgID:      profile.OrgID.String,
	}
	renderJSON(w, r, http.StatusOK, resp, logger)
}

// getStateByID returns an "account state" for the given ID.
func getStateByID(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		instrumentation.GetAccountStateError()
		renderPlain(w, r, http.StatusBadRequest, "unable to assert X-Rh-Identity header", logger)
		return
	}
	logger = log.With().Interface("identity", id).Logger()

	profileID := chi.URLParam(r, "id")
	if profileID == "" {
		instrumentation.GetAccountStateError()
		renderPlain(w, r, http.StatusBadRequest, "cannot get ID from URL", logger)
		return
	}

	profile, err := db.GetProfile(profileID)
	if err != nil {
		instrumentation.GetAccountStateError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get profile for ID: %v", err), logger)
		return
	}

	resp := stateArchive{
		Account:   profile.AccountID.String,
		ID:        profile.ID.String(),
		Label:     profile.Label.String,
		Initiator: profile.Creator.String,
		CreatedAt: profile.CreatedAt.UTC(),
		State:     profile.StateConfig(),
		OrgID:     profile.OrgID.String,
	}
	renderJSON(w, r, http.StatusOK, resp, logger)
}

// postManage sets the value of the apply_state field on the current account
// state record.
func postManage(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		renderPlain(w, r, http.StatusBadRequest, "unable to assert X-Rh-Identity header", logger)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("cannot read request body: %v", err), logger)
		return
	}

	var enabled bool
	if err := json.Unmarshal(data, &enabled); err != nil {
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("cannot unmarshal json: %v", err), logger)
		return
	}

	logger.Info().Bool("apply_state", enabled).Msg("setting apply_state for account")

	currentProfile, err := db.GetCurrentProfile(id.Identity.OrgID)
	if err != nil {
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get current profile: %v", err), logger)
		return
	}

	newProfile := db.CopyProfile(*currentProfile)
	newProfile.Active = enabled

	if err := db.InsertProfile(newProfile); err != nil {
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot insert profile: %v", err), logger)
		return
	}

	renderNone(w, r, http.StatusOK, logger)
}

func getStatesIDPlaybook(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		renderPlain(w, r, http.StatusBadRequest, "unable to assert X-Rh-Identity header", logger)
		return
	}
	logger = logger.With().Interface("identity", id).Logger()

	profileID := chi.URLParam(r, "id")
	if profileID == "" {
		instrumentation.PlaybookRequestError()
		renderPlain(w, r, http.StatusBadRequest, "cannot get ID from URL", logger)
		return
	}

	profile, err := db.GetProfile(profileID)
	if err != nil {
		instrumentation.PlaybookRequestError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot get profile for ID: %v", err), logger)
		return
	}

	playbook, err := internal.GeneratePlaybook(profile.StateConfig())
	if err != nil {
		instrumentation.PlaybookRequestError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot generate playbook: %v", err), logger)
		return
	}

	instrumentation.PlaybookRequestOK()
	renderPlain(w, r, http.StatusOK, playbook, logger)
}

// postStatesPreview renders a playbook with the contents of the HTTP request
// body as input.
func postStatesPreview(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Str("path", r.URL.Path).Str("method", r.Method).Logger()

	id, ok := r.Context().Value(identity.Key).(identity.XRHID)
	if !ok {
		renderPlain(w, r, http.StatusBadRequest, "unable to assert X-Rh-Identity header", logger)
		return
	}
	logger = logger.With().Interface("identity", id).Logger()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		instrumentation.PlaybookRequestError()
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("cannot read request body: %v", err), logger)
		return
	}

	var stateConfig map[string]string
	if err := json.Unmarshal(data, &stateConfig); err != nil {
		instrumentation.PlaybookRequestError()
		renderPlain(w, r, http.StatusBadRequest, fmt.Sprintf("cannot unmarshal request body: %v", err), logger)
		return
	}

	playbook, err := internal.GeneratePlaybook(stateConfig)
	if err != nil {
		instrumentation.PlaybookRequestError()
		renderPlain(w, r, http.StatusInternalServerError, fmt.Sprintf("cannot generate playbook: %v", err), logger)
		return
	}

	renderPlain(w, r, http.StatusOK, playbook, logger)
}
