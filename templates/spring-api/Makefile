.PHONY: up down run test image full

IMAGE ?= starter-api:local

ifeq ($(OS),Windows_NT)
MVNW ?= .\mvnw.cmd
else
MVNW ?= ./mvnw
endif

# forge:if db!=none||cache=redis
up: ## sobe a infraestrutura (banco e cache)
	docker compose up -d --wait

down: ## derruba a infraestrutura (use "docker compose down -v" para apagar o volume)
	docker compose down

run: up ## sobe a infraestrutura e roda a aplicação com o perfil local
# forge:else
#~ run: ## roda a aplicação com o perfil local
# forge:end
	$(MVNW) spring-boot:run -Dspring-boot.run.profiles=local

test: ## testes, Spotless, ArchUnit e JaCoCo (precisa de Docker)
	$(MVNW) verify

image: ## constrói a imagem Docker
	docker build -t $(IMAGE) .

full: ## sobe a infraestrutura e a aplicação em containers
# forge:if db!=none||cache=redis
	docker compose -f compose.yaml -f compose.full.yaml up --build
# forge:else
#~ 	docker compose -f compose.full.yaml up --build
# forge:end
