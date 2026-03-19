BEGIN;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS follower_auth_bindings (
    follower_id UUID PRIMARY KEY,
    device_token_hash CHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_follower_auth_bindings_updated_at ON follower_auth_bindings;
CREATE TRIGGER trg_follower_auth_bindings_updated_at
BEFORE UPDATE ON follower_auth_bindings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMIT;
