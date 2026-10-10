package com.example.starterapi.adapter.out.client;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.starterapi.domain.port.out.AddressLookupPort;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.test.context.TestPropertySource;

@TestPropertySource(properties = "clients.viacep.impl=webclient")
class ViaCepWebClientTest extends AbstractViaCepClientTest {

  @Autowired AddressLookupPort lookupPort;

  @org.junit.jupiter.api.Test
  void usesWebClientAdapter() {
    assertThat(lookupPort).isInstanceOf(ViaCepWebClientAdapter.class);
  }
}
