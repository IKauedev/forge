package com.example.starterapi.config;

import com.example.starterapi.domain.port.out.AddressLookupPort;
import com.example.starterapi.domain.port.out.AddressRepositoryPort;
import com.example.starterapi.domain.service.AddressLookupService;
import jakarta.enterprise.inject.Produces;
import jakarta.enterprise.inject.Typed;

/** O domínio não conhece o CDI: seus beans são registrados aqui. */
public class DomainConfig {

  /**
   * {@code @Typed} limita o tipo do bean: sem ele, {@code AddressLookupService} (que implementa o
   * caso de uso) concorreria com {@code AddressApplicationService} na injeção de {@code
   * LookupAddressUseCase}.
   */
  @Produces
  @Typed(AddressLookupService.class)
  AddressLookupService addressLookupService(
      AddressLookupPort lookupPort, AddressRepositoryPort repositoryPort) {
    return new AddressLookupService(lookupPort, repositoryPort);
  }

  // forge:if db=none
  // ~ @Produces
  // ~ AddressRepositoryPort addressRepository() {
  // ~   return new NoOpAddressRepository();
  // ~ }
  // forge:end
}
