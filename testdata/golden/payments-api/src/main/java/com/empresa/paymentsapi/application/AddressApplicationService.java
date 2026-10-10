package com.empresa.paymentsapi.application;

import com.empresa.paymentsapi.domain.model.Address;
import com.empresa.paymentsapi.domain.port.in.LookupAddressUseCase;
import com.empresa.paymentsapi.domain.service.AddressLookupService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Service;

/** Camada de aplicação: orquestra o caso de uso do domínio e cuida de preocupações transversais. */
@Service
@Primary
public class AddressApplicationService implements LookupAddressUseCase {

  private static final Logger log = LoggerFactory.getLogger(AddressApplicationService.class);

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
