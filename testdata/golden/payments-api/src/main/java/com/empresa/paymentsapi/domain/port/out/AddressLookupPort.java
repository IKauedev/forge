package com.empresa.paymentsapi.domain.port.out;

import com.empresa.paymentsapi.domain.model.Address;
import java.util.Optional;

/** Consulta de endereço em um provedor externo. */
public interface AddressLookupPort {

  Optional<Address> findByCep(String cep);
}
