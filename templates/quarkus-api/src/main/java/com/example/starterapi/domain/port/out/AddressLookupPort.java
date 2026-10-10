package com.example.starterapi.domain.port.out;

import com.example.starterapi.domain.model.Address;
import java.util.Optional;

/** Consulta de endereço em um provedor externo. */
public interface AddressLookupPort {

  Optional<Address> findByCep(String cep);
}
