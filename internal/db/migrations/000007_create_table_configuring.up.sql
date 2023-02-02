CREATE TABLE IF NOT EXISTS configuring (
    host_id UUID NOT NULL,
    profile_id UUID NOT NULL,
    CONSTRAINT fk_profile_id FOREIGN KEY (profile_id) REFERENCES profiles(profile_id),
    CONSTRAINT configuring_pkey PRIMARY KEY (host_id, profile_id)
);
