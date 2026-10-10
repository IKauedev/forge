package com.empresa.ordersservice.config;

import com.empresa.ordersservice.domain.port.out.AddressLookupPort;
import com.empresa.ordersservice.domain.port.out.AddressRepositoryPort;
import com.empresa.ordersservice.domain.service.AddressLookupService;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.resilience.annotation.EnableResilientMethods;

/** O domínio não conhece o Spring: seus beans são registrados aqui. */
@Configuration
@EnableResilientMethods
public class DomainConfig {

  @Bean
  AddressLookupService addressLookupService(
      AddressLookupPort lookupPort, AddressRepositoryPort repositoryPort) {
    return new AddressLookupService(lookupPort, repositoryPort);
  }
}
