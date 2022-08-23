BEGIN;

-- Enable the pgcrypto extension
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create the profiles table
CREATE TABLE IF NOT EXISTS profiles (
    profile_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT,
    label TEXT,
    account_id TEXT,
    org_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    creator TEXT,
    insights BOOLEAN NOT NULL DEFAULT FALSE,
    remediations BOOLEAN NOT NULL DEFAULT FALSE,
    compliance BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS org_id_idx ON profiles (org_id);

-- Insert records into profiles using existing data from state_archive.
INSERT INTO
    profiles (
        profile_id,
        label,
        account_id,
        created_at,
        creator,
        insights,
        remediations,
        compliance
    )
SELECT
    state_id AS profile_id,
    label,
    account_id,
    created_at,
    initiator AS creator,
    state ->> 'insights' = 'enabled' AS insights,
    state ->> 'remediations' = 'enabled' AS remediations,
    state ->> 'compliance_openscap' = 'enabled' AS compliance
FROM
    state_archive;

-- Insert records into profiles using existing data from account_states.
INSERT INTO
    profiles (
        profile_id,
        label,
        account_id,
        active,
        insights,
        remediations,
        compliance
    )
SELECT
    state_id AS profile_id,
    label,
    account_id,
    apply_state AS active,
    state ->> 'insights' = 'enabled' AS insights,
    state ->> 'remediations' = 'enabled' AS remediations,
    state ->> 'compliance_openscap' = 'enabled' AS compliance
FROM
    account_states ON CONFLICT (profile_id) DO
UPDATE
SET
    active = (
        SELECT
            apply_state
        FROM
            account_states
        WHERE
            state_id = EXCLUDED.profile_id
    );

DROP TABLE account_states;

DROP TABLE state_archive;

COMMIT;
