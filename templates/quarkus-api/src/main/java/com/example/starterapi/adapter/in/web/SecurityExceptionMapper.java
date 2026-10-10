package com.example.starterapi.adapter.in.web;

import io.quarkus.security.AuthenticationFailedException;
import io.quarkus.security.ForbiddenException;
import io.quarkus.security.UnauthorizedException;
import jakarta.ws.rs.core.HttpHeaders;
import org.jboss.resteasy.reactive.RestResponse;
import org.jboss.resteasy.reactive.server.ServerExceptionMapper;

/** Erros 401 e 403 no mesmo formato (application/problem+json) dos demais erros da API. */
public class SecurityExceptionMapper {

  private static final String CHALLENGE = "Bearer";

  /** Sem credencial: o endpoint exige autenticação. */
  @ServerExceptionMapper
  public RestResponse<ProblemDetail> unauthorized(UnauthorizedException ex) {
    return unauthorized("Não autenticado", "Informe um token Bearer válido.");
  }

  /** Token presente, mas inválido ou vencido. */
  @ServerExceptionMapper
  public RestResponse<ProblemDetail> authenticationFailed(AuthenticationFailedException ex) {
    return unauthorized("Token inválido", "O token é inválido ou expirou. Entre novamente.");
  }

  /** Autenticado, mas sem a role exigida pelo endpoint. */
  @ServerExceptionMapper
  public RestResponse<ProblemDetail> forbidden(ForbiddenException ex) {
    return ProblemDetail.response(
        403, "Acesso negado", "Você não tem permissão para este recurso.");
  }

  private static RestResponse<ProblemDetail> unauthorized(String title, String detail) {
    return RestResponse.ResponseBuilder.<ProblemDetail>create(RestResponse.Status.UNAUTHORIZED)
        .header(HttpHeaders.WWW_AUTHENTICATE, CHALLENGE)
        .entity(ProblemDetail.of(401, title, detail))
        .type(ProblemDetail.MEDIA_TYPE)
        .build();
  }
}
