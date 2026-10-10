package com.empresa.slimservice.domain.port.in;

import com.empresa.slimservice.domain.model.Address;

public interface LookupAddressUseCase {

  /**
   * @throws com.empresa.slimservice.domain.exception.InvalidCepException se o CEP não tiver 8
   *     dígitos
   * @throws com.empresa.slimservice.domain.exception.AddressNotFoundException se o CEP não existir
   */
  Address lookup(String cep);
}
