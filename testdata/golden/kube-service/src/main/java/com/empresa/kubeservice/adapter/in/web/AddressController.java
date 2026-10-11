package com.empresa.kubeservice.adapter.in.web;

import com.empresa.kubeservice.domain.port.in.LookupAddressUseCase;
import com.empresa.kubeservice.shared.security.IsAuthenticated;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@IsAuthenticated
@RequestMapping("/api/addresses")
public class AddressController {
  private final LookupAddressUseCase lookupAddress;

  public AddressController(LookupAddressUseCase lookupAddress) {
    this.lookupAddress = lookupAddress;
  }

  @GetMapping("/{cep}")
  public AddressResponse get(@PathVariable String cep) {
    return AddressResponse.from(lookupAddress.lookup(cep));
  }
}
