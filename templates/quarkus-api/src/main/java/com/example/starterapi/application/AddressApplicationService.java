package com.example.starterapi.application;

import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.in.LookupAddressUseCase;
import com.example.starterapi.domain.service.AddressLookupService;
import jakarta.enterprise.context.ApplicationScoped;
import org.jboss.logging.Logger;

/** Camada de aplicação: orquestra o caso de uso do domínio e cuida de preocupações transversais. */
@ApplicationScoped
public class AddressApplicationService implements LookupAddressUseCase {

  private static final Logger log = Logger.getLogger(AddressApplicationService.class);

  private final AddressLookupService domainService;

  public AddressApplicationService(AddressLookupService domainService) {
    this.domainService = domainService;
  }

  @Override
  public Address lookup(String cep) {
    log.info("Consultando endereço");
    return domainService.lookup(cep);
  }
}
