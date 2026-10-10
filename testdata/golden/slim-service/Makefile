.PHONY: up down run test image full

IMAGE ?= slim-service:local

ifeq ($(OS),Windows_NT)
MVNW ?= .\mvnw.cmd
else
MVNW ?= ./mvnw
endif

run: ## roda a aplicação com o perfil local
	$(MVNW) spring-boot:run -Dspring-boot.run.profiles=local

test: ## testes, Spotless, ArchUnit e JaCoCo (precisa de Docker)
	$(MVNW) verify

image: ## constrói a imagem Docker
	docker build -t $(IMAGE) .

full: ## sobe a infraestrutura e a aplicação em containers
	docker compose -f compose.full.yaml up --build
