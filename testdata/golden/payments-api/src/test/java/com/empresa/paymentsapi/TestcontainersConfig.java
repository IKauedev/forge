package com.empresa.paymentsapi;

import org.springframework.boot.test.context.TestConfiguration;
import org.springframework.boot.testcontainers.service.connection.ServiceConnection;
import org.springframework.context.annotation.Bean;

@TestConfiguration(proxyBeanMethods = false)
public class TestcontainersConfig {
  @Bean
  @ServiceConnection
  org.testcontainers.postgresql.PostgreSQLContainer database() {
    return new org.testcontainers.postgresql.PostgreSQLContainer("postgres:17-alpine");
  }

  // Alternativas fora do estado de referência usam nomes totalmente qualificados, porque o
  // Spotless remove imports sem uso no template.
}
