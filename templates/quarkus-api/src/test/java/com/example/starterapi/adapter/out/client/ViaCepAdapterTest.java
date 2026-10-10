package com.example.starterapi.adapter.out.client;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.example.starterapi.domain.exception.ExternalServiceException;
import com.example.starterapi.domain.port.out.AddressLookupPort;
import com.example.starterapi.shared.CorrelationId;
import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import jakarta.inject.Inject;
import org.jboss.logging.MDC;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

@QuarkusTest
@QuarkusTestResource(value = ViaCepStub.class, restrictToAnnotatedClass = true)
class ViaCepAdapterTest {

  @Inject AddressLookupPort adapter;

  @BeforeEach
  void reset() {
    ViaCepStub.calls.set(0);
    ViaCepStub.correlationIds.clear();
    MDC.remove(CorrelationId.MDC_KEY);
  }

  @Test
  void mapsTheProviderContractToTheDomain() {
    var address = adapter.findByCep("01001000").orElseThrow();

    assertThat(address.cep()).isEqualTo("01001000"); // só dígitos
    assertThat(address.street()).isEqualTo("Praça da Sé");
    assertThat(address.neighborhood()).isEqualTo("Sé");
    assertThat(address.city()).isEqualTo("São Paulo");
    assertThat(address.state()).isEqualTo("SP");
  }

  @Test
  void unknownCepIsEmpty() {
    assertThat(adapter.findByCep("99999999")).isEmpty();
  }

  @Test
  void badRequestFromTheProviderIsEmptyAndIsNotRetried() {
    assertThat(adapter.findByCep("00000000")).isEmpty();
    assertThat(ViaCepStub.calls.get()).isEqualTo(1);
  }

  @Test
  void serverErrorIsRetriedThenBecomesAnExternalServiceException() {
    assertThatThrownBy(() -> adapter.findByCep("50000000"))
        .isInstanceOf(ExternalServiceException.class);

    assertThat(ViaCepStub.calls.get()).isEqualTo(3); // 1 chamada + 2 retries
  }

  @Test
  void forwardsTheCorrelationIdToTheProvider() {
    MDC.put(CorrelationId.MDC_KEY, "corr-42");

    adapter.findByCep("01001000");

    assertThat(ViaCepStub.correlationIds).containsExactly("corr-42");
  }
}
