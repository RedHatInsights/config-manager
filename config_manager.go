package main

import (
	"config-manager/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/spf13/viper"
)

// ConfigManager struct
type ConfigManager struct {
	Router *mux.Router
	DB     *sql.DB
	Config *viper.Viper
}

// Init establishes the database connection and mux
func (cm *ConfigManager) Init() {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		cm.Config.GetString("DBUser"),
		cm.Config.GetString("DBPass"),
		cm.Config.GetString("DBName"))

	var err error
	cm.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	cm.Router = mux.NewRouter()
	cm.routes()
}

// Run starts the http server for ConfigManager
func (cm *ConfigManager) Run(addr string) {
	utils.StartHTTPServer(addr, "config-manager", cm.Router)
}

func (cm *ConfigManager) routes() {
	s := cm.Router.PathPrefix("/configmanager").Subrouter()
	s.Use(identity.EnforceIdentity)
	s.HandleFunc("/", cm.getAccountState).Methods("GET")
	s.HandleFunc("/update", cm.updateAccountState).Methods("PUT")
	s.HandleFunc("/apply", cm.applyAccountState).Methods("POST")
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

func (cm *ConfigManager) getAccountState(w http.ResponseWriter, r *http.Request) {
	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Println("Getting state for account: ", id.Identity.AccountNumber)
	acc := &account{AccountID: id.Identity.AccountNumber}
	if err := acc.getAccount(cm.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			fmt.Println("Creating new account entry")
			cm.createAccountEntry(w, acc)
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, acc)
}

func (cm *ConfigManager) createAccountEntry(w http.ResponseWriter, acc *account) {
	acc.State = StateMap{
		"insights":   "enabled",
		"advisor":    "enabled",
		"compliance": "enabled",
	}

	if err := acc.createAccount(cm.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusCreated, acc)
}

func (cm *ConfigManager) updateAccountState(w http.ResponseWriter, r *http.Request) {
	var id identity.XRHID
	id = identity.Get(r.Context())
	fmt.Println("Updating state for account: ", id.Identity.AccountNumber)

	var state StateMap
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&state); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	acc := account{
		AccountID: id.Identity.AccountNumber,
		State:     state,
	}
	if err := acc.updateAccount(cm.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, acc)
}

func (cm *ConfigManager) applyAccountState(w http.ResponseWriter, r *http.Request) {
	fmt.Println("This is where a work request will be sent to the playbook dispatcher")
	fmt.Println("This is also where creating a new entry for the 'run' table will be initiated")
}
