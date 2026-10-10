package com.empresa.ordersservice.adapter.out.client;

import com.empresa.ordersservice.domain.exception.ExternalServiceException;
import com.empresa.ordersservice.domain.model.Address;
import com.empresa.ordersservice.domain.port.out.AddressLookupPort;
import java.time.Duration;
import java.util.Optional;
import org.springframework.resilience.annotation.Retryable;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.web.reactive.function.client.WebClientRequestException;
import org.springframework.web.reactive.function.client.WebClientResponseException;

public class ViaCepWebClientAdapter implements AddressLookupPort {

  private final WebClient webClient;
  private final ViaCepMapper mapper;
  private final Duration timeout;

  public ViaCepWebClientAdapter(WebClient webClient, ViaCepMapper mapper, Duration timeout) {
    this.webClient = webClient;
    this.mapper = mapper;
    this.timeout = timeout;
  }

  @Override
  @Retryable(
      includes = ExternalServiceException.class,
      maxRetriesString = "${clients.viacep.retry.max-retries:2}",
      delayString = "${clients.viacep.retry.delay-ms:200}")
  public Optional<Address> findByCep(String cep) {
    try {
      ViaCepResponse response =
          webClient
              .get()
              .uri("/{cep}/json", cep)
              .retrieve()
              .bodyToMono(ViaCepResponse.class)
              .block(timeout);
      if (response == null || Boolean.TRUE.equals(response.erro())) {
        return Optional.empty();
      }
      return Optional.of(mapper.toDomain(response));
    } catch (WebClientResponseException e) {
      if (e.getStatusCode().value() == 400 || e.getStatusCode().value() == 404) {
        return Optional.empty();
      }
      throw new ExternalServiceException("Falha ao consultar ViaCEP (webclient)", e);
    } catch (WebClientRequestException e) {
      throw new ExternalServiceException("Falha ao consultar ViaCEP (webclient)", e);
    }
  }
}
