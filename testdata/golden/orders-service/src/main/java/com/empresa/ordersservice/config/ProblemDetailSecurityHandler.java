package com.empresa.ordersservice.config;

import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.security.access.AccessDeniedException;
import org.springframework.security.core.AuthenticationException;
import org.springframework.security.web.AuthenticationEntryPoint;
import org.springframework.security.web.access.AccessDeniedHandler;

/** Responde 401/403 no mesmo formato (RFC 9457) dos demais erros da API. */
class ProblemDetailSecurityHandler implements AuthenticationEntryPoint, AccessDeniedHandler {

  @Override
  public void commence(
      HttpServletRequest request, HttpServletResponse response, AuthenticationException ex)
      throws IOException {
    write(
        response,
        HttpStatus.UNAUTHORIZED,
        "Não autenticado",
        "Token ausente, inválido ou expirado.");
  }

  @Override
  public void handle(
      HttpServletRequest request, HttpServletResponse response, AccessDeniedException ex)
      throws IOException {
    write(
        response,
        HttpStatus.FORBIDDEN,
        "Acesso negado",
        "Você não tem permissão para este recurso.");
  }

  private static void write(
      HttpServletResponse response, HttpStatus status, String title, String detail)
      throws IOException {
    response.setStatus(status.value());
    response.setContentType(MediaType.APPLICATION_PROBLEM_JSON_VALUE);
    response.setCharacterEncoding(StandardCharsets.UTF_8.name());
    response
        .getWriter()
        .write(
            "{\"type\":\"about:blank\",\"title\":\"%s\",\"status\":%d,\"detail\":\"%s\"}"
                .formatted(title, status.value(), detail));
  }
}
