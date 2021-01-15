CREATE TABLE IF NOT EXISTS accounts
(
    account_id TEXT NOT NULL,
    state JSONB NOT NULL,
    CONSTRAINT accounts_pkey PRIMARY KEY (account_id)
);
