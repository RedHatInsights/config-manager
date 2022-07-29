#!/bin/bash

TOKEN=${TOKEN:-undefined}
GABI_HOST=${GABI_HOST:-localhost}
DB_HOST=${DB_HOST:-localhost}
DB_USER=${DB_USER:-insights}
DB_PASS=${DB_PASS:-insights}
DB_NAME=${DB_NAME:-insights}
DB_PORT=${DB_PORT:-5432}
DB_URN="postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo '{"query":"SELECT * FROM profiles;"}' | 
    ht GET "https://${GABI_HOST}/query" "Authorization: Bearer ${TOKEN}" |
    jq -c .result[1:][] |
    spyql -Otable=profiles "SELECT json[0] AS profile_id, json[1] AS name, json[2] AS label, json[3] AS account_id, json[4] AS org_id, json[5] AS created_at, json[6] AS active, json[7] AS creator, json[8] AS insights, json[9] AS remediations, json[10] AS compliance FROM json to sql" |
    psql "${DB_URN}"
