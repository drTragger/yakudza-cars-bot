# Load environment variables from the .env file
include .env
export $(shell sed 's/=.*//' .env)

# Місце, де зберігаються файли міграцій
MIGRATION_DIR=./db/migrations

# Команда для створення нової міграції
create-migration:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $$name

# Команда для запуску всіх міграцій
migrate-up:
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up

# Команда для скасування останньої міграції
migrate-down:
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down 1

# Команда для скасування всіх міграцій
migrate-drop:
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" drop

# Команда для перегляду версії міграцій
migrate-version:
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" version

# Команда для встановлення golang-migrate, якщо ще не встановлений
install-migrate:
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest