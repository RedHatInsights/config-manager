#!/bin/bash

TOKEN=${TOKEN:-undefined}
GABI_HOST=${GABI_HOST:-localhost}
DB_HOST=${DB_HOST:-localhost}
DB_USER=${DB_USER:-insights}
DB_PASS=${DB_PASS:-insights}
DB_NAME=${DB_NAME:-insights}
DB_PORT=${DB_PORT:-5432}
DB_URN="postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/insights?sslmode=disable"

echo '{"query":"SELECT * FROM account_states;"}' | 
    ht GET "https://${GABI_HOST}/query" "Authorization: Bearer ${TOKEN}" |
    jq -c .result[1:][] | 
    spyql -Otable=account_states "SELECT json[0] AS account_id, json[1] AS state, json[2] AS state_id, json[3] AS label, json[4] AS apply_state FROM json TO sql" |
    psql "${DB_URN}"

echo '{"query":"SELECT * FROM state_archive;"}' |
    ht GET "https://${GABI_HOST}/query" "Authorization: Bearer ${TOKEN}" |
    jq -c .result[1:][] |
    spyql -Otable=state_archive "SELECT json[0] AS state_id, json[1] AS account_id, json[2] AS label, json[3] AS initiator, json[4] AS created_at, json[5] AS state FROM json TO sql" |
    psql "${DB_URN}"
