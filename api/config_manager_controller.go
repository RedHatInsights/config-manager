package api

import (
	"config-manager/application"
	"config-manager/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redhatinsights/platform-go-middlewares/identity"
)

type ConfigManagerController struct {
	ConfigManagerService *application.ConfigManagerService
	Router               *mux.Router
}

func (cmc *ConfigManagerController) Run(addr string) {
	utils.StartHTTPServer(addr, "config-manager", cmc.Router)
}

func (cmc *ConfigManagerController) Routes() {
	s := cmc.Router.PathPrefix("/config").Subrouter()
	s.Use(identity.EnforceIdentity)
	s.HandleFunc("/state", cmc.getAccountState).Methods("GET")
	s.HandleFunc("/state", cmc.updateAccountState).Methods("POST")
	s.HandleFunc("/state/apply", cmc.applyAccountState).Methods("POST")
	s.HandleFunc("/state/changes", cmc.getStateChanges).Methods("GET") // Add sorting/limit/pagination
	s.HandleFunc("/runs/{label}", cmc.getRuns).Methods("GET")          // Add sorting/limit/pagination
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	res, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, map[string]string{"error": message})
}

func (cmc *ConfigManagerController) getAccountState(w http.ResponseWriter, r *http.Request) {
	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Println("Getting state for account: ", id.Identity.AccountNumber)

	acc, err := cmc.ConfigManagerService.GetAccountState(id.Identity.AccountNumber)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, acc)
}

func (cmc *ConfigManagerController) updateAccountState(w http.ResponseWriter, r *http.Request) {
	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Println("Updating state for account: ", id.Identity.AccountNumber)

	var payload map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	acc, err := cmc.ConfigManagerService.UpdateAccountState(id.Identity.AccountNumber, "demo-user", payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	//TODO get number of connected clients (inventory repository?) and build response object that contains
	//both acc and number of clients for 'pre flight check' purposes

	respondWithJSON(w, http.StatusOK, acc)
}

func (cmc *ConfigManagerController) applyAccountState(w http.ResponseWriter, r *http.Request) {
	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Println("Applying state for account: ", id.Identity.AccountNumber)

	clients, err := cmc.ConfigManagerService.GetClients(id.Identity.AccountNumber)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	run, err := cmc.ConfigManagerService.ApplyState(id.Identity.AccountNumber, "demo-user", clients.Clients)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, run)
}

func (cmc *ConfigManagerController) getStateChanges(w http.ResponseWriter, r *http.Request) {
	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Println("Getting state changes for account: ", id.Identity.AccountNumber)

	states, err := cmc.ConfigManagerService.GetStateChanges(id.Identity.AccountNumber, 3, 0) // Add limit, offset, and sort-by to query params
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, states)
}

func (cmc *ConfigManagerController) getRuns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	label := vars["label"]

	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Printf("Getting runs for account %s with label %s", id.Identity.AccountNumber, label)

	runs, err := cmc.ConfigManagerService.GetRunsByLabel(label, 3, 0) // Add limit, offset, and sort-by to query params
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, runs)
}

func (cmc *ConfigManagerController) getRunStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["runID"]

	// var id identity.XRHID
	// id = identity.Get(r.Context())
	fmt.Println("Getting status for run: ", runID)

	run, err := cmc.ConfigManagerService.GetSingleRun(runID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	status, err := cmc.ConfigManagerService.GetRunStatus(run.Label)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, status)
}
