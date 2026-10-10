package com.example.starterapi;

import static io.restassured.RestAssured.given;
import static org.assertj.core.api.Assertions.assertThat;
import static org.hamcrest.Matchers.containsString;
import static org.hamcrest.Matchers.equalTo;

import com.example.starterapi.adapter.out.client.ViaCepStub;
import com.example.starterapi.shared.security.Roles;
import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.security.TestSecurity;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/**
 * Fluxo de ponta a ponta com tudo real: recurso REST, caso de uso, domínio, adapter do ViaCEP
 * (contra um provedor falso), banco (Dev Services) e o mapeamento de erros. Sem mocks do caso de
 * uso.
 */
@QuarkusTest
@QuarkusTestResource(value = ViaCepStub.class, restrictToAnnotatedClass = true)
// forge:if auth!=none
@TestSecurity(user = "ana", roles = Roles.USER)
// forge:end
class AddressFlowTest {

  @BeforeEach
  void reset() {
    ViaCepStub.calls.set(0);
  }

  @Test
  void looksTheAddressUpAtTheProviderAndAnswersWithTheDomainData() {
    given()
        .get("/api/addresses/01001-000")
        .then()
        .statusCode(200)
        .body("cep", equalTo("01001000"))
        .body("street", equalTo("Praça da Sé"))
        .body("city", equalTo("São Paulo"))
        .body("state", equalTo("SP"));
  }

  @Test
  void aKnownAddressIsServedWithoutCallingTheProviderAgain() {
    given().get("/api/addresses/01001000").then().statusCode(200);
    ViaCepStub.calls.set(0);

    given().get("/api/addresses/01001000").then().statusCode(200);

    // forge:if db!=none
    assertThat(ViaCepStub.calls.get()).as("a segunda consulta sai do banco").isZero();
    // forge:else
    // ~ assertThat(ViaCepStub.calls.get()).as("sem banco consulta de novo").isEqualTo(1);
    // forge:end
  }

  @Test
  void anUnknownCepIs404() {
    given()
        .get("/api/addresses/99999999")
        .then()
        .statusCode(404)
        .contentType(containsString("application/problem+json"))
        .body("title", equalTo("Endereço não encontrado"));
  }

  @Test
  void aMalformedCepIs400WithoutCallingTheProvider() {
    given().get("/api/addresses/123").then().statusCode(400).body("title", equalTo("CEP inválido"));

    assertThat(ViaCepStub.calls.get()).isZero();
  }

  @Test
  void aProviderOutageBecomesA503WithoutLeakingDetails() {
    given()
        .get("/api/addresses/50000000")
        .then()
        .statusCode(503)
        .body("title", equalTo("Serviço externo indisponível"));

    assertThat(ViaCepStub.calls.get()).as("1 chamada + 2 retries").isEqualTo(3);
  }
}
