package com.example.starterapi.domain.port.out;

import com.example.starterapi.domain.model.Address;
import java.util.Optional;

/** Armazenamento local de endereços já consultados. */
public interface AddressRepositoryPort {

  Optional<Address> findByCep(String cep);

  void save(Address address);
}
