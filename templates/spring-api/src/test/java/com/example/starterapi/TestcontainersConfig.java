package com.example.starterapi;

import org.springframework.boot.test.context.TestConfiguration;
import org.springframework.boot.testcontainers.service.connection.ServiceConnection;
import org.springframework.context.annotation.Bean;

@TestConfiguration(proxyBeanMethods = false)
public class TestcontainersConfig {

  // forge:if db!=none
  // forge:slot db-test-container
  @Bean
  @ServiceConnection
  org.testcontainers.postgresql.PostgreSQLContainer database() {
    return new org.testcontainers.postgresql.PostgreSQLContainer("postgres:17-alpine");
  }
  // forge:end
  // forge:end

  // Alternativas fora do estado de referência usam nomes totalmente qualificados, porque o
  // Spotless remove imports sem uso no template.
  // forge:if cache=redis
  // ~ @Bean
  // ~ @ServiceConnection(name = "redis")
  // ~ org.testcontainers.containers.GenericContainer<?> redis() {
  // ~   return new org.testcontainers.containers.GenericContainer<>("redis:8-alpine")
  // ~       .withExposedPorts(6379);
  // ~ }
  // forge:end
}
