package com.empresa.ordersservice.domain.exception;

public class AddressNotFoundException extends RuntimeException {

  public AddressNotFoundException(String cep) {
    super("Endereço não encontrado para o CEP " + cep);
  }
}
