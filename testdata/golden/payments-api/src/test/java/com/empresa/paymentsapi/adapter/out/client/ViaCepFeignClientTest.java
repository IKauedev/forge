package com.empresa.paymentsapi.adapter.out.client;

import static org.assertj.core.api.Assertions.assertThat;

import com.empresa.paymentsapi.domain.port.out.AddressLookupPort;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.test.context.TestPropertySource;

@TestPropertySource(properties = "clients.viacep.impl=feign")
class ViaCepFeignClientTest extends AbstractViaCepClientTest {

  @Autowired AddressLookupPort lookupPort;

  @org.junit.jupiter.api.Test
  void usesFeignAdapter() {
    assertThat(lookupPort).isInstanceOf(ViaCepFeignAdapter.class);
  }
}
