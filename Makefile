.PHONY: up down container restart build run clear
up:
	docker compose up || docker-compose up
down:
	docker compose down || docker-compose down
container:
	docker exec -it simple-golang-chat-app sh
restart:
	make down && make up && docker volume prune -f
build:
	go build -C ./cmd -o ../bin/chat
run:
	./bin/chat
clear:
	rm -rf ./bin/*
