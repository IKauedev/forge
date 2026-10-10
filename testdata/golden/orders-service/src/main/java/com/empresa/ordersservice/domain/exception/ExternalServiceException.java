package com.empresa.ordersservice.domain.exception;

/** Falha transitória ao chamar um serviço externo; candidata a retry. */
public class ExternalServiceException extends RuntimeException {

  public ExternalServiceException(String message, Throwable cause) {
    super(message, cause);
  }
}
