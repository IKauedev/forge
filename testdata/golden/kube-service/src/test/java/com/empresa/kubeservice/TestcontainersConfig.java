package com.empresa.kubeservice;

import org.springframework.boot.test.context.TestConfiguration;
import org.springframework.boot.testcontainers.service.connection.ServiceConnection;
import org.springframework.context.annotation.Bean;

@TestConfiguration(proxyBeanMethods = false)
public class TestcontainersConfig {
  @Bean
  @ServiceConnection
  org.testcontainers.mysql.MySQLContainer database() {
    return new org.testcontainers.mysql.MySQLContainer("mysql:8.4");
  }

  // Alternativas fora do estado de referência usam nomes totalmente qualificados, porque o
  // Spotless remove imports sem uso no template.
  @Bean
  @ServiceConnection(name = "redis")
  org.testcontainers.containers.GenericContainer<?> redis() {
    return new org.testcontainers.containers.GenericContainer<>("redis:8-alpine")
        .withExposedPorts(6379);
  }
}
