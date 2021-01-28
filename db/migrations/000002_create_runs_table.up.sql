CREATE TABLE IF NOT EXISTS runs
(
    run_id UUID NOT NULL,
    account_id TEXT NOT NULL,
    hostname TEXT NOT NULL,
    initiator TEXT NOT NULL,
    label TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT run_pkey PRIMARY KEY (run_id)
);
