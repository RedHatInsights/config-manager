CREATE TABLE IF NOT EXISTS state_archive
(
    state_id UUID NOT NULL,
    account_id TEXT NOT NULL,
    label TEXT NOT NULL,
    initiator TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    state JSONB NOT NULL,
    CONSTRAINT state_archive_pkey PRIMARY KEY (state_id)
);
