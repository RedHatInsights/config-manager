CREATE TABLE IF NOT EXISTS account_states
(
    account_id TEXT NOT NULL,
    state JSONB NOT NULL,
    state_id UUID NOT NULL,
    label TEXT NOT NULL,
    CONSTRAINT accounts_pkey PRIMARY KEY (account_id)
);
