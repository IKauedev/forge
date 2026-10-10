package com.example.starterapi.adapter.out.client;

import com.example.starterapi.shared.CorrelationId;
import jakarta.ws.rs.client.ClientRequestContext;
import jakarta.ws.rs.client.ClientRequestFilter;
import org.jboss.logging.MDC;

/** Repassa o id de correlação da requisição atual para a chamada ao provedor externo. */
public class CorrelationIdClientFilter implements ClientRequestFilter {

  @Override
  public void filter(ClientRequestContext context) {
    Object id = MDC.get(CorrelationId.MDC_KEY);
    if (id != null) {
      context.getHeaders().putSingle(CorrelationId.HEADER, id.toString());
    }
  }
}
