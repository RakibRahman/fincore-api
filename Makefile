# Load environment variables from .env file
include .env
export

# PostgreSQL data directory for version 18.x
POSTGRES_DATA_DIR := /var/lib/postgresql/18/main

# Start Postgres
postgres:
	docker volume create $(POSTGRES_VOLUME)
	docker run --name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		-v $(POSTGRES_VOLUME):$(POSTGRES_DATA_DIR) \
		-d postgres:18.1

# View logs
logs:
	docker logs $(POSTGRES_CONTAINER)

# Reset database completely (container + data)
reset:
	docker rm -f $(POSTGRES_CONTAINER)
	docker volume rm -f $(POSTGRES_VOLUME)


# Create database (if not exists)
createdb:
	docker exec -it $(POSTGRES_CONTAINER) createdb --username=$(POSTGRES_USER) --owner=$(POSTGRES_USER) $(POSTGRES_DB)

# Drop database
dropdb:
	docker exec -it $(POSTGRES_CONTAINER) dropdb $(POSTGRES_DB)

# Stop and remove Postgres container
stop-postgres:
	-@docker stop $(POSTGRES_CONTAINER)

migrate-db-up:
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" -verbose up

migrate-db-down:
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" -verbose down

sqlc:
	sqlc generate