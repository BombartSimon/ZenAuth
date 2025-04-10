CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  email TEXT,
  is_external BOOLEAN DEFAULT false
);

-- Index pour recherche rapide d'utilisateurs par username
CREATE INDEX idx_users_username ON users(username);
-- Index pour recherche d'utilisateurs par email (si utilisé dans vos requêtes)
CREATE INDEX idx_users_email ON users(email);

CREATE TABLE auth_codes (
  code TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  redirect_uri TEXT NOT NULL,
  user_id UUID NOT NULL,
  code_challenge TEXT,
  code_challenge_method TEXT,
  expires_at TIMESTAMP NOT NULL,
  scope TEXT NOT NULL
);

-- Index pour améliorer les performances des jointures user_id sans contrainte FK
CREATE INDEX idx_auth_codes_user_id ON auth_codes(user_id);
-- Index sur client_id pour accélération des requêtes filtrées par client
CREATE INDEX idx_auth_codes_client_id ON auth_codes(client_id);
-- Index pour le nettoyage des codes expirés
CREATE INDEX idx_auth_codes_expires_at ON auth_codes(expires_at);

CREATE TABLE clients (
  id TEXT PRIMARY KEY,
  secret TEXT NOT NULL,
  name TEXT NOT NULL,
  redirect_uris TEXT[] DEFAULT '{}'
);

-- Index pour recherche de clients par nom (si utilisé)
CREATE INDEX idx_clients_name ON clients(name);

-- Exemple de client
INSERT INTO clients (id, secret, name, redirect_uris)
VALUES ('demo-client', 'demo-secret', 'Demo App', ARRAY['http://localhost:3000']);

CREATE TABLE refresh_tokens (
  token TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  user_id TEXT,
  issued_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth_providers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    tenant_id TEXT,
    enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Run this SQL to add new columns to the users table

ALTER TABLE users ADD COLUMN IF NOT EXISTS external_id TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider TEXT;
CREATE INDEX IF NOT EXISTS idx_users_external_id ON users(external_id);

-- Index pour recherche par client_id
CREATE INDEX idx_refresh_tokens_client_id ON refresh_tokens(client_id);
-- Index pour recherche par user_id (sans FK)
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
-- Index pour expiration/nettoyage des tokens anciens
CREATE INDEX idx_refresh_tokens_issued_at ON refresh_tokens(issued_at);

-- Extensions utiles
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- pour gen_random_uuid()