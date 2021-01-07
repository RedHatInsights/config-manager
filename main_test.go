package main

import (
	"config-manager/config"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var cm ConfigManager

func TestMain(m *testing.M) {
	cm.Config = config.Get()
	cm.Init()

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func TestNewAccount(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/configmanager/", nil)
	req.Header.Set("x-rh-identity", "eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=")
	response := executeRequest(req)

	checkResponse(t, http.StatusCreated, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	cm.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponse(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code to be %d. Got %d\n", expected, actual)
	}
}

func ensureTableExists() {
	if _, err := cm.DB.Exec(accountsTableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	cm.DB.Exec("DELETE FROM accounts")
	//cm.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const hostsTableCreationQuery = `CREATE TABLE IF NOT EXISTS hosts
(
    client_id TEXT NOT NULL,
    inventory_id TEXT NOT NULL,
	account TEXT NOT NULL,
	run_id TEXT,
	state JSON NOT NULL,
    CONSTRAINT hosts_pkey PRIMARY KEY (client_id)
)`

const accountsTableCreationQuery = `CREATE TABLE IF NOT EXISTS accounts
(
	account_id TEXT NOT NULL,
	state JSONB NOT NULL,
	CONSTRAINT accounts_pkey PRIMARY KEY (account_id)
)`
