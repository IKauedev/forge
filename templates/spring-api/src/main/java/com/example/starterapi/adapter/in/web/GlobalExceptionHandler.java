package com.example.starterapi.adapter.in.web;

import com.example.starterapi.domain.exception.AddressNotFoundException;
import com.example.starterapi.domain.exception.ExternalServiceException;
import com.example.starterapi.domain.exception.InvalidCepException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ProblemDetail;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.servlet.mvc.method.annotation.ResponseEntityExceptionHandler;

@RestControllerAdvice
public class GlobalExceptionHandler extends ResponseEntityExceptionHandler {

  @ExceptionHandler(InvalidCepException.class)
  ProblemDetail invalidCep(InvalidCepException ex) {
    return problem(HttpStatus.BAD_REQUEST, "CEP inválido", ex.getMessage());
  }

  @ExceptionHandler(AddressNotFoundException.class)
  ProblemDetail notFound(AddressNotFoundException ex) {
    return problem(HttpStatus.NOT_FOUND, "Endereço não encontrado", ex.getMessage());
  }

  @ExceptionHandler(ExternalServiceException.class)
  ProblemDetail unavailable(ExternalServiceException ex) {
    return problem(
        HttpStatus.SERVICE_UNAVAILABLE,
        "Serviço externo indisponível",
        "Não foi possível consultar o provedor de CEP. Tente novamente.");
  }

  private static ProblemDetail problem(HttpStatus status, String title, String detail) {
    ProblemDetail pd = ProblemDetail.forStatusAndDetail(status, detail);
    pd.setTitle(title);
    return pd;
  }
}
