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
	ConfigManagerService application.ConfigManagerService
	Router               *mux.Router
}

func (cmc *ConfigManagerController) Init() {
	cmc.Router = mux.NewRouter()
	cmc.routes()
}

func (cmc *ConfigManagerController) Run(addr string) {
	utils.StartHTTPServer(addr, "config-manager", cmc.Router)
}

func (cmc *ConfigManagerController) routes() {
	s := cmc.Router.PathPrefix("/configmanager").Subrouter()
	s.Use(identity.EnforceIdentity)
	s.HandleFunc("/state", cmc.getAccountState).Methods("GET")
	s.HandleFunc("/state", cmc.updateAccountState).Methods("POST")
	s.HandleFunc("/sync", cmc.applyAccountState).Methods("POST")
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

	acc, err := cmc.ConfigManagerService.GetAccount(id.Identity.AccountNumber)
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

	acc, err := cmc.ConfigManagerService.UpdateAccount(id.Identity.AccountNumber, payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, acc)
}

func (cmc *ConfigManagerController) applyAccountState(w http.ResponseWriter, r *http.Request) {
	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Println("Applying state for account: ", id.Identity.AccountNumber)

	// TODO: Add handler for sending work / creating run entry
}
