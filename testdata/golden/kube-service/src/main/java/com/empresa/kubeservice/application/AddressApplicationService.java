package com.empresa.kubeservice.application;

import com.empresa.kubeservice.domain.model.Address;
import com.empresa.kubeservice.domain.port.in.LookupAddressUseCase;
import com.empresa.kubeservice.domain.service.AddressLookupService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.cache.annotation.Cacheable;
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
  @Cacheable("addresses")
  public Address lookup(String cep) {
    log.info("Consultando endereço");
    return domainService.lookup(cep);
  }
}
