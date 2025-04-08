CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL
);

CREATE TABLE auth_codes (
  code TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  redirect_uri TEXT NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id),
  code_challenge TEXT,
  code_challenge_method TEXT,
  expires_at TIMESTAMP NOT NULL,
  scope TEXT NOT NULL

);

CREATE TABLE clients (
  id TEXT PRIMARY KEY,
  secret TEXT NOT NULL,
  name TEXT NOT NULL,
  redirect_uris TEXT[] DEFAULT '{}'
);

-- Exemple de client
INSERT INTO clients (id, secret, name, redirect_uris)
VALUES ('demo-client', 'demo-secret', 'Demo App', ARRAY['http://localhost:3000']);

CREATE TABLE refresh_tokens (
  token TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  user_id TEXT,
  issued_at TIMESTAMP NOT NULL DEFAULT now()
);



-- Extensions utiles
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- pour gen_random_uuid()
