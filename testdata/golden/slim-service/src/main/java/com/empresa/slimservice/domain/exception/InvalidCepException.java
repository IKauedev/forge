package com.empresa.slimservice.domain.exception;

public class InvalidCepException extends RuntimeException {

  public InvalidCepException(String cep) {
    super("CEP inválido: " + cep);
  }
}
