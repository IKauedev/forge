package com.example.starterapi.config;

import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.out.AddressRepositoryPort;
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
