package com.empresa.paymentsapi.domain.port.in;

import com.empresa.paymentsapi.domain.model.Address;

public interface LookupAddressUseCase {

  /**
   * @throws com.empresa.paymentsapi.domain.exception.InvalidCepException se o CEP não tiver 8
   *     dígitos
   * @throws com.empresa.paymentsapi.domain.exception.AddressNotFoundException se o CEP não existir
   */
  Address lookup(String cep);
}
