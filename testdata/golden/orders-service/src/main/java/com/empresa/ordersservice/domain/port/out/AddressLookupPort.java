package com.empresa.ordersservice.domain.port.out;

import com.empresa.ordersservice.domain.model.Address;
import java.util.Optional;

/** Consulta de endereço em um provedor externo. */
public interface AddressLookupPort {

  Optional<Address> findByCep(String cep);
}
