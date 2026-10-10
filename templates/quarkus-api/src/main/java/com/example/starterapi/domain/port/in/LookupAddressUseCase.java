package com.example.starterapi.domain.port.in;

import com.example.starterapi.domain.model.Address;

public interface LookupAddressUseCase {

  /**
   * Devolve o endereço do CEP. O nome das exceções é simples de propósito: um nome totalmente
   * qualificado mudaria de tamanho com o nome do projeto e o Spotless reformataria o comentário.
   *
   * @throws InvalidCepException se o CEP não tiver 8 dígitos
   * @throws AddressNotFoundException se o CEP não existir
   */
  Address lookup(String cep);
}
