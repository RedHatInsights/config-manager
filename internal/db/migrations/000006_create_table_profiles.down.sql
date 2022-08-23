BEGIN;

CREATE TABLE IF NOT EXISTS account_states (
    account_id TEXT NOT NULL,
    state JSONB NOT NULL,
    state_id UUID NOT NULL,
    label TEXT NOT NULL,
    apply_state BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT accounts_pkey PRIMARY KEY (account_id)
);

CREATE TABLE IF NOT EXISTS state_archive (
    state_id UUID NOT NULL,
    account_id TEXT NOT NULL,
    label TEXT NOT NULL,
    initiator TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    state JSONB NOT NULL,
    CONSTRAINT state_archive_pkey PRIMARY KEY (state_id)
);

-- Insert records into state_archive using data from the profiles table. This
-- query reconstructs the state_archive table using all profiles rows.
INSERT INTO
    state_archive (
        account_id,
        state,
        state_id,
        label,
        created_at,
        initiator
    )
SELECT
    account_id,
    jsonb_build_object(
        'insights',
        CASE
            WHEN insights = TRUE THEN 'enabled'
            ELSE 'disabled'
        END,
        'remediations',
        CASE
            WHEN insights = TRUE THEN 'enabled'
            ELSE 'disabled'
        END,
        'compliance_openscap',
        CASE
            WHEN compliance = TRUE THEN 'enabled'
            ELSE 'disabled'
        END
    ) AS state,
    profile_id AS state_id,
    label,
    created_at,
    COALESCE(creator, 'redhat') AS initiator
FROM
    profiles
ORDER BY
    profiles.account_id,
    profiles.created_at DESC;

-- Insert records into account_states using data from the profiles table. This
-- query reconstructs the account_states table using the most recent profile for
-- each account ID.
INSERT INTO
    account_states (account_id, state, state_id, label, apply_state)
SELECT
    DISTINCT ON (account_id) account_id,
    jsonb_build_object(
        'insights',
        CASE
            WHEN insights = TRUE THEN 'enabled'
            ELSE 'disabled'
        END,
        'remediations',
        CASE
            WHEN insights = TRUE THEN 'enabled'
            ELSE 'disabled'
        END,
        'compliance_openscap',
        CASE
            WHEN compliance = TRUE THEN 'enabled'
            ELSE 'disabled'
        END
    ) AS state,
    profile_id AS state_id,
    label,
    active AS apply_state
FROM
    profiles
ORDER BY
    profiles.account_id,
    profiles.created_at DESC ON CONFLICT (account_id) DO NOTHING;

DROP INDEX IF EXISTS org_id_idx;

DROP TABLE profiles;

COMMIT;
