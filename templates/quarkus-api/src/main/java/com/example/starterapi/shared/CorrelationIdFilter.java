package com.example.starterapi.shared;

import io.quarkus.vertx.http.runtime.filters.Filters;
import jakarta.enterprise.event.Observes;
import jakarta.ws.rs.container.ContainerRequestContext;
import java.util.UUID;
import org.jboss.logging.MDC;
import org.jboss.resteasy.reactive.server.ServerRequestFilter;
import org.jboss.resteasy.reactive.server.ServerResponseFilter;

/**
 * Garante o X-Correlation-Id em toda requisição: reaproveita o que veio no cabeçalho ou gera um,
 * põe no MDC dos logs e devolve na resposta.
 *
 * <p>São duas etapas porque o MDC é por thread: o filtro do Vert.x roda antes da autenticação
 * (assim até um 401 sai com o cabeçalho), e o filtro JAX-RS roda na mesma thread do endpoint.
 */
public class CorrelationIdFilter {

  void register(@Observes Filters filters) {
    filters.register(
        rc -> {
          String id = rc.request().getHeader(CorrelationId.HEADER);
          if (id == null || id.isBlank()) {
            id = UUID.randomUUID().toString();
            rc.request().headers().set(CorrelationId.HEADER, id);
          }
          rc.response().putHeader(CorrelationId.HEADER, id);
          rc.next();
        },
        1100);
  }

  @ServerRequestFilter
  public void putInMdc(ContainerRequestContext request) {
    MDC.put(CorrelationId.MDC_KEY, request.getHeaderString(CorrelationId.HEADER));
  }

  @ServerResponseFilter
  public void clearMdc() {
    MDC.remove(CorrelationId.MDC_KEY);
  }
}
