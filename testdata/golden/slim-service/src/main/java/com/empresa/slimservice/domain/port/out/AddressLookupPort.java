package com.empresa.slimservice.domain.port.out;

import com.empresa.slimservice.domain.model.Address;
import java.util.Optional;

/** Consulta de endereço em um provedor externo. */
public interface AddressLookupPort {

  Optional<Address> findByCep(String cep);
}
