APP_NAME=messenger

DB_URL=postgres://user:password@localhost:5432/test?sslmode=disable

MIGRATIONS_PATH=./migrations


run:
	go mod tidy
	go run cmd/messenger/main.go




up:
	docker compose up -d

down:
	docker compose down

down-v:
	docker compose down -v

logs:
	docker compose logs -f




migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

migrate-down-all:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down



help:
	@echo "make run"
	@echo "make up"
	@echo "make down"
	@echo "make migrate-create name=init"
	@echo "make migrate-up"
	@echo "make migrate-down"
