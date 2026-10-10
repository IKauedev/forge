package com.example.starterapi.adapter.in.web;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.containsString;
import static org.hamcrest.Matchers.equalTo;
import static org.hamcrest.Matchers.matchesPattern;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import com.example.starterapi.domain.exception.AddressNotFoundException;
import com.example.starterapi.domain.exception.ExternalServiceException;
import com.example.starterapi.domain.exception.InvalidCepException;
import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.in.LookupAddressUseCase;
import com.example.starterapi.shared.CorrelationId;
import com.example.starterapi.shared.security.Roles;
import io.quarkus.test.InjectMock;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.security.TestSecurity;
import org.junit.jupiter.api.Test;

@QuarkusTest
// forge:if auth!=none
@TestSecurity(user = "ana", roles = Roles.USER)
// forge:end
class AddressResourceTest {

  @InjectMock LookupAddressUseCase lookup;

  @Test
  void returnsTheAddress() {
    when(lookup.lookup("01001000"))
        .thenReturn(new Address("01001000", "Praça da Sé", "lado ímpar", "Sé", "São Paulo", "SP"));

    given()
        .get("/api/addresses/01001000")
        .then()
        .statusCode(200)
        .body("cep", equalTo("01001000"))
        .body("street", equalTo("Praça da Sé"))
        .body("state", equalTo("SP"));
  }

  @Test
  void invalidCepIsAProblemDetailWith400() {
    when(lookup.lookup(anyString())).thenThrow(new InvalidCepException("123"));

    given()
        .get("/api/addresses/123")
        .then()
        .statusCode(400)
        .contentType(containsString("application/problem+json"))
        .body("title", equalTo("CEP inválido"))
        .body("status", equalTo(400));
  }

  @Test
  void unknownCepIs404() {
    when(lookup.lookup(anyString())).thenThrow(new AddressNotFoundException("99999999"));

    given().get("/api/addresses/99999999").then().statusCode(404);
  }

  @Test
  void providerFailureIs503WithoutLeakingDetails() {
    when(lookup.lookup(anyString()))
        .thenThrow(new ExternalServiceException("boom interno", new RuntimeException()));

    given()
        .get("/api/addresses/01001000")
        .then()
        .statusCode(503)
        .body("detail", equalTo("Não foi possível consultar o provedor de CEP. Tente novamente."));
  }

  @Test
  void everyResponseCarriesACorrelationId() {
    when(lookup.lookup(anyString()))
        .thenReturn(new Address("01001000", "Praça da Sé", "", "Sé", "São Paulo", "SP"));

    given()
        .get("/api/addresses/01001000")
        .then()
        .header(CorrelationId.HEADER, matchesPattern("[0-9a-f-]{36}"));
  }

  @Test
  void reusesTheCorrelationIdSentByTheCaller() {
    when(lookup.lookup(anyString()))
        .thenReturn(new Address("01001000", "Praça da Sé", "", "Sé", "São Paulo", "SP"));

    given()
        .header(CorrelationId.HEADER, "abc-123")
        .get("/api/addresses/01001000")
        .then()
        .header(CorrelationId.HEADER, equalTo("abc-123"));
  }
}
