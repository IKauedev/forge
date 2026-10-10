package com.example.starterapi.adapter.in.web;

import com.example.starterapi.domain.exception.AddressNotFoundException;
import com.example.starterapi.domain.exception.ExternalServiceException;
import com.example.starterapi.domain.exception.InvalidCepException;
import org.eclipse.microprofile.faulttolerance.exceptions.CircuitBreakerOpenException;
import org.jboss.resteasy.reactive.RestResponse;
import org.jboss.resteasy.reactive.server.ServerExceptionMapper;

public class GlobalExceptionMapper {

  @ServerExceptionMapper
  public RestResponse<ProblemDetail> invalidCep(InvalidCepException ex) {
    return ProblemDetail.response(400, "CEP inválido", ex.getMessage());
  }

  @ServerExceptionMapper
  public RestResponse<ProblemDetail> notFound(AddressNotFoundException ex) {
    return ProblemDetail.response(404, "Endereço não encontrado", ex.getMessage());
  }

  @ServerExceptionMapper
  public RestResponse<ProblemDetail> unavailable(ExternalServiceException ex) {
    return unavailable();
  }

  /** O circuit breaker aberto significa que o provedor está fora do ar há pouco tempo. */
  @ServerExceptionMapper
  public RestResponse<ProblemDetail> circuitOpen(CircuitBreakerOpenException ex) {
    return unavailable();
  }

  private static RestResponse<ProblemDetail> unavailable() {
    return ProblemDetail.response(
        503,
        "Serviço externo indisponível",
        "Não foi possível consultar o provedor de CEP. Tente novamente.");
  }
}
