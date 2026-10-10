.PHONY: dev up down test image full

IMAGE ?= starter-api:local

ifeq ($(OS),Windows_NT)
MVNW ?= .\mvnw.cmd
else
MVNW ?= ./mvnw
endif

dev: ## modo dev do Quarkus (live reload; o Dev Services sobe o banco descartável, precisa de Docker)
	$(MVNW) quarkus:dev

# forge:if db!=none
up: ## sobe um banco fixo (compose.yaml), no lugar do Dev Services
	docker compose up -d --wait

down: ## derruba o banco (use "docker compose down -v" para apagar o volume)
	docker compose down

# forge:end
test: ## testes, Spotless e ArchUnit (precisa de Docker)
	$(MVNW) verify

image: ## constrói a imagem Docker
	docker build -t $(IMAGE) .

full: ## sobe o banco e a aplicação em containers
# forge:if db!=none
	docker compose -f compose.yaml -f compose.full.yaml up --build
# forge:else
#~ 	docker compose -f compose.full.yaml up --build
# forge:end
