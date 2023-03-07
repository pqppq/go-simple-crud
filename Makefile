all: build run
build:
	docker build --tag go-simple-crud-api .
run:
	docker compose --file ./docker-compose.yaml up
clean:
	docker compose --file ./docker-compose.yaml down
