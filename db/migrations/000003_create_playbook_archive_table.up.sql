CREATE TABLE IF NOT EXISTS playbook_archive
(
    playbook_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    state JSONB NOT NULL,
    CONSTRAINT playbook_archive_pkey PRIMARY KEY (playbook_id)
);