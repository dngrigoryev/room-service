.PHONY: up seed down test

up:
	docker-compose up -d --build

seed:
	docker-compose exec -T postgres psql -U pguser -d room_booking < testdata/seed.sql

down:
	docker-compose down

test:
	go test -v ./internal/service/...
	go test -v ./tests/...