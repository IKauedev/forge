package com.empresa.slimservice.config;

import com.empresa.slimservice.domain.model.Address;
import com.empresa.slimservice.domain.port.out.AddressRepositoryPort;
import java.util.Optional;

/** Repositório sem persistência: usado quando o projeto é gerado com --db none. */
class NoOpAddressRepository implements AddressRepositoryPort {

  @Override
  public Optional<Address> findByCep(String cep) {
    return Optional.empty();
  }

  @Override
  public void save(Address address) {
    // sem banco, nada a guardar
  }
}
