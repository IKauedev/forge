package com.empresa.slimservice.adapter.out.client;

import com.empresa.slimservice.domain.exception.ExternalServiceException;
import com.empresa.slimservice.domain.model.Address;
import com.empresa.slimservice.domain.port.out.AddressLookupPort;
import feign.FeignException;
import java.util.Optional;
import org.springframework.resilience.annotation.Retryable;

public class ViaCepFeignAdapter implements AddressLookupPort {

  private final ViaCepFeignClient client;
  private final ViaCepMapper mapper;

  public ViaCepFeignAdapter(ViaCepFeignClient client, ViaCepMapper mapper) {
    this.client = client;
    this.mapper = mapper;
  }

  @Override
  @Retryable(
      includes = ExternalServiceException.class,
      maxRetriesString = "${clients.viacep.retry.max-retries:2}",
      delayString = "${clients.viacep.retry.delay-ms:200}")
  public Optional<Address> findByCep(String cep) {
    try {
      ViaCepResponse response = client.find(cep);
      if (response == null || Boolean.TRUE.equals(response.erro())) {
        return Optional.empty();
      }
      return Optional.of(mapper.toDomain(response));
    } catch (FeignException e) {
      if (e.status() == 400 || e.status() == 404) {
        return Optional.empty();
      }
      throw new ExternalServiceException("Falha ao consultar ViaCEP (feign)", e);
    }
  }
}
