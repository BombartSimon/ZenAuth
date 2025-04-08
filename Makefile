# Variables
DB_CONTAINER=oauth-db
DB_USER=oauth_user
DB_NAME=oauth
DB_PASSWORD=oauth_pass
SEED_USER=demo
SEED_PASS=demo123
REDIRECT_PORT=3000

.PHONY: all up init-db seed run server client logs down

# Lance tout
all: up init-db seed run client

# Démarre PostgreSQL via Docker Compose
up:
	docker-compose up -d

# Attendre que PostgreSQL soit prêt et exécuter le script SQL
init-db:
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@until docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) > /dev/null 2>&1; do sleep 1; done
	@echo "✅ PostgreSQL is ready. Running init.sql..."
	@docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < init.sql || true

# Générer un utilisateur avec mot de passe hashé
seed:
	@echo "🔑 Seeding user: $(SEED_USER)"
	@HASH=$$(go run ./scripts/gen_hash.go $(SEED_PASS)) && \
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c \
	"INSERT INTO users (username, password_hash) VALUES ('$(SEED_USER)', '$$HASH') ON CONFLICT DO NOTHING;"
	@echo "✅ User '$(SEED_USER)' seeded with password '$(SEED_PASS)'"

# Démarre le serveur ZenAuth
run:
	@echo "🚀 Starting ZenAuth server..."
	go run cmd/main.go

# Lance le client web (nécessite python3)
client:
	@echo "🌐 Starting OAuth client on http://localhost:$(REDIRECT_PORT)"
	cd client-test && python3 -m http.server $(REDIRECT_PORT)

# Affiche les logs du container PostgreSQL
logs:
	docker-compose logs -f

# Arrête tout
down:
	docker-compose down
