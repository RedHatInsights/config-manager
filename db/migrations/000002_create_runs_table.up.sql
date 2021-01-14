CREATE TABLE IF NOT EXISTS runs
(
	run_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    initiator TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
	CONSTRAINT run_pkey PRIMARY KEY (run_id)
);