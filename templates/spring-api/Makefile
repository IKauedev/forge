.PHONY: up down run test image full

IMAGE ?= starter-api:local

ifeq ($(OS),Windows_NT)
MVNW ?= .\mvnw.cmd
else
MVNW ?= ./mvnw
endif

up: ## sobe a infraestrutura (Postgres)
	docker compose up -d --wait

down: ## derruba a infraestrutura (use "docker compose down -v" para apagar o volume)
	docker compose down

run: up ## sobe o banco e roda a aplicação com o perfil local
	$(MVNW) spring-boot:run -Dspring-boot.run.profiles=local

test: ## testes, Spotless, ArchUnit e JaCoCo (precisa de Docker)
	$(MVNW) verify

image: ## constrói a imagem Docker
	docker build -t $(IMAGE) .

full: ## sobe banco + aplicação em containers
	docker compose -f compose.yaml -f compose.full.yaml up --build
