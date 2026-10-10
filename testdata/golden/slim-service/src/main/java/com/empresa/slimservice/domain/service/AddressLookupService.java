package com.empresa.slimservice.domain.service;

import com.empresa.slimservice.domain.exception.AddressNotFoundException;
import com.empresa.slimservice.domain.exception.InvalidCepException;
import com.empresa.slimservice.domain.model.Address;
import com.empresa.slimservice.domain.port.in.LookupAddressUseCase;
import com.empresa.slimservice.domain.port.out.AddressLookupPort;
import com.empresa.slimservice.domain.port.out.AddressRepositoryPort;

public class AddressLookupService implements LookupAddressUseCase {

  private final AddressLookupPort lookupPort;
  private final AddressRepositoryPort repositoryPort;

  public AddressLookupService(AddressLookupPort lookupPort, AddressRepositoryPort repositoryPort) {
    this.lookupPort = lookupPort;
    this.repositoryPort = repositoryPort;
  }

  @Override
  public Address lookup(String rawCep) {
    String cep = normalize(rawCep);
    return repositoryPort
        .findByCep(cep)
        .or(() -> lookupPort.findByCep(cep).map(this::store))
        .orElseThrow(() -> new AddressNotFoundException(cep));
  }

  private Address store(Address address) {
    repositoryPort.save(address);
    return address;
  }

  static String normalize(String rawCep) {
    String digits = rawCep == null ? "" : rawCep.replaceAll("\\D", "");
    if (digits.length() != 8) {
      throw new InvalidCepException(rawCep);
    }
    return digits;
  }
}
