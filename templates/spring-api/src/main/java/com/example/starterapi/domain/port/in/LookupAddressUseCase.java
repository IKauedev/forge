package com.example.starterapi.domain.port.in;

import com.example.starterapi.domain.model.Address;

public interface LookupAddressUseCase {

  /**
   * @throws com.example.starterapi.domain.exception.InvalidCepException se o CEP não tiver 8
   *     dígitos
   * @throws com.example.starterapi.domain.exception.AddressNotFoundException se o CEP não existir
   */
  Address lookup(String cep);
}
