
run:
	go run cmd/main.go

up-db:
	docker-compose -f postgres/docker-compose.yml build
	docker-compose -f postgres/docker-compose.yml up -d

down-db:
	docker-compose -f postgres/docker-compose.yml down
