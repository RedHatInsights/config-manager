CREATE TABLE IF NOT EXISTS playbook_archive
(
    account_id TEXT NOT NULL,
    state_id UUID NOT NULL,
    playbook TEXT NOT NULL,
    CONSTRAINT playbook_archive_pkey PRIMARY KEY (state_id)
);