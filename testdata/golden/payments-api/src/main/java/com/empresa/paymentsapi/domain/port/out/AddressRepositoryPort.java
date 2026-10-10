package com.empresa.paymentsapi.domain.port.out;

import com.empresa.paymentsapi.domain.model.Address;
import java.util.Optional;

/** Armazenamento local de endereços já consultados. */
public interface AddressRepositoryPort {

  Optional<Address> findByCep(String cep);

  void save(Address address);
}
