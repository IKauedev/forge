package com.empresa.ordersservice.domain.port.in;

import com.empresa.ordersservice.domain.model.Address;

public interface LookupAddressUseCase {

  /**
   * @throws com.empresa.ordersservice.domain.exception.InvalidCepException se o CEP não tiver 8
   *     dígitos
   * @throws com.empresa.ordersservice.domain.exception.AddressNotFoundException se o CEP não existir
   */
  Address lookup(String cep);
}
