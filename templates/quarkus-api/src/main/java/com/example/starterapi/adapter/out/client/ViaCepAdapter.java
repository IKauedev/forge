package com.example.starterapi.adapter.out.client;

import com.example.starterapi.domain.exception.ExternalServiceException;
import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.out.AddressLookupPort;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.ws.rs.ProcessingException;
import jakarta.ws.rs.WebApplicationException;
import java.time.temporal.ChronoUnit;
import java.util.Optional;
import org.eclipse.microprofile.faulttolerance.CircuitBreaker;
import org.eclipse.microprofile.faulttolerance.Retry;
import org.eclipse.microprofile.rest.client.inject.RestClient;

/**
 * Resiliência com SmallRye Fault Tolerance: tenta de novo em falhas transitórias e abre o circuito
 * quando o provedor está fora do ar. Os valores podem ser trocados em application.properties, ex.:
 * {@code <classe>/findByCep/Retry/maxRetries=5}.
 */
@ApplicationScoped
public class ViaCepAdapter implements AddressLookupPort {

  private final ViaCepRestClient client;

  public ViaCepAdapter(@RestClient ViaCepRestClient client) {
    this.client = client;
  }

  @Override
  @Retry(
      retryOn = ExternalServiceException.class,
      maxRetries = 2,
      delay = 200,
      delayUnit = ChronoUnit.MILLIS)
  @CircuitBreaker(
      failOn = ExternalServiceException.class,
      requestVolumeThreshold = 10,
      failureRatio = 0.5,
      delay = 10,
      delayUnit = ChronoUnit.SECONDS)
  public Optional<Address> findByCep(String cep) {
    try {
      ViaCepResponse response = client.find(cep);
      if (response == null || Boolean.TRUE.equals(response.erro())) {
        return Optional.empty();
      }
      return Optional.of(ViaCepMapper.toDomain(response));
    } catch (WebApplicationException e) {
      int status = e.getResponse().getStatus();
      if (status == 400 || status == 404) {
        return Optional.empty();
      }
      throw new ExternalServiceException("Falha ao consultar ViaCEP (HTTP " + status + ")", e);
    } catch (ProcessingException e) {
      throw new ExternalServiceException("Falha ao consultar ViaCEP (rede)", e);
    }
  }
}
