GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOVULNCHECK ?= govulncheck

.PHONY: up down container restart build run clear lint vulncheck test ci
up:
	docker compose up || docker-compose up
down:
	docker compose down || docker-compose down
container:
	docker exec -it simple-golang-chat-app sh
restart:
	make down && make up && docker volume prune -f
build:
	$(GO) build -C ./cmd -o ../bin/chat
lint:
	$(GOLANGCI_LINT) run
vulncheck:
	$(GOVULNCHECK) ./...
test:
	$(GO) test -race ./...
ci: lint vulncheck test
run:
	./bin/chat
clear:
	rm -rf ./bin/*
