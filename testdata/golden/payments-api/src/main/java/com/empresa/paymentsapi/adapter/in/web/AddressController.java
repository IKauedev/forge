package com.empresa.paymentsapi.adapter.in.web;

import com.empresa.paymentsapi.domain.port.in.LookupAddressUseCase;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
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
