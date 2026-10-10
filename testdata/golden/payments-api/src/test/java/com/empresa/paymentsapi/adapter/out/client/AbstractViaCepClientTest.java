package com.empresa.paymentsapi.adapter.out.client;

import static com.github.tomakehurst.wiremock.client.WireMock.aResponse;
import static com.github.tomakehurst.wiremock.client.WireMock.equalTo;
import static com.github.tomakehurst.wiremock.client.WireMock.getRequestedFor;
import static com.github.tomakehurst.wiremock.client.WireMock.urlEqualTo;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.header;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.empresa.paymentsapi.TestcontainersConfig;
import com.empresa.paymentsapi.shared.CorrelationId;
import com.github.tomakehurst.wiremock.WireMockServer;
import com.github.tomakehurst.wiremock.client.WireMock;
import com.github.tomakehurst.wiremock.core.WireMockConfiguration;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.webmvc.test.autoconfigure.AutoConfigureMockMvc;
import org.springframework.context.annotation.Import;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.springframework.test.web.servlet.MockMvc;

/** Cenários comuns às duas implementações do cliente ViaCEP (feign e webclient). */
@SpringBootTest
@AutoConfigureMockMvc
@ActiveProfiles("test")
@Import(TestcontainersConfig.class)
abstract class AbstractViaCepClientTest {

  static WireMockServer wireMock;

  @Autowired MockMvc mvc;

  @BeforeAll
  static void startWireMock() {
    wireMock = new WireMockServer(WireMockConfiguration.options().dynamicPort());
    wireMock.start();
  }

  @AfterAll
  static void stopWireMock() {
    wireMock.stop();
  }

  @DynamicPropertySource
  static void props(DynamicPropertyRegistry registry) {
    if (wireMock == null) {
      startWireMock();
    }
    registry.add("clients.viacep.url", wireMock::baseUrl);
  }

  @BeforeEach
  void reset() {
    wireMock.resetAll();
  }

  @Test
  void fetchesAddressAndPropagatesCorrelationId() throws Exception {
    wireMock.stubFor(
        WireMock.get(urlEqualTo("/01001000/json"))
            .willReturn(
                aResponse()
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        """
                        {"cep":"01001-000","logradouro":"Praça da Sé","complemento":"lado ímpar",
                         "bairro":"Sé","localidade":"São Paulo","uf":"SP"}
                        """)));

    mvc.perform(get("/api/addresses/01001-000").with(jwt()).header(CorrelationId.HEADER, "abc-123"))
        .andExpect(status().isOk())
        .andExpect(header().string(CorrelationId.HEADER, "abc-123"))
        .andExpect(jsonPath("$.cep").value("01001000"))
        .andExpect(jsonPath("$.street").value("Praça da Sé"))
        .andExpect(jsonPath("$.state").value("SP"));

    wireMock.verify(
        getRequestedFor(urlEqualTo("/01001000/json"))
            .withHeader(CorrelationId.HEADER, equalTo("abc-123")));
  }

  @Test
  void unknownCepIs404() throws Exception {
    wireMock.stubFor(
        WireMock.get(urlEqualTo("/99999999/json"))
            .willReturn(
                aResponse()
                    .withHeader("Content-Type", "application/json")
                    .withBody("{\"erro\":true}")));

    mvc.perform(get("/api/addresses/99999999").with(jwt())).andExpect(status().isNotFound());
  }

  @Test
  void serverErrorIsRetriedThenReturns503() throws Exception {
    wireMock.stubFor(
        WireMock.get(urlEqualTo("/30130010/json")).willReturn(aResponse().withStatus(500)));

    mvc.perform(get("/api/addresses/30130010").with(jwt()))
        .andExpect(status().isServiceUnavailable());

    // 1 tentativa + 2 retries (clients.viacep.retry.max-retries=2)
    wireMock.verify(3, getRequestedFor(urlEqualTo("/30130010/json")));
  }

  @Test
  void secondRequestIsServedFromDatabase() throws Exception {
    wireMock.stubFor(
        WireMock.get(urlEqualTo("/20040020/json"))
            .willReturn(
                aResponse()
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        """
                        {"cep":"20040-020","logradouro":"Rua da Assembleia","complemento":"",
                         "bairro":"Centro","localidade":"Rio de Janeiro","uf":"RJ"}
                        """)));

    mvc.perform(get("/api/addresses/20040020").with(jwt())).andExpect(status().isOk());
    mvc.perform(get("/api/addresses/20040020").with(jwt())).andExpect(status().isOk());

    wireMock.verify(1, getRequestedFor(urlEqualTo("/20040020/json")));
  }
}
