# syntax=docker/dockerfile:1

# ---- build: compila e gera o fast-jar em camadas ----
FROM eclipse-temurin:25-jdk-alpine AS builder
WORKDIR /builder

# Camada de dependências: só muda quando o pom.xml muda
COPY .mvn/ .mvn/
COPY mvnw pom.xml ./
RUN --mount=type=cache,target=/root/.m2 ./mvnw -q -B dependency:go-offline

COPY src/ src/
RUN --mount=type=cache,target=/root/.m2 ./mvnw -q -B -DskipTests package

# ---- runtime: só JRE, usuário não-root ----
FROM eclipse-temurin:25-jre-alpine
WORKDIR /app

RUN addgroup -S app && adduser -S -G app app

# Da camada que menos muda para a que mais muda
COPY --from=builder /builder/target/quarkus-app/lib/ ./lib/
COPY --from=builder /builder/target/quarkus-app/quarkus/ ./quarkus/
COPY --from=builder /builder/target/quarkus-app/app/ ./app/
COPY --from=builder /builder/target/quarkus-app/quarkus-run.jar ./

USER app
EXPOSE 8080

ENTRYPOINT ["java", "-XX:MaxRAMPercentage=75.0", "-XX:+ExitOnOutOfMemoryError", "-jar", "quarkus-run.jar"]
