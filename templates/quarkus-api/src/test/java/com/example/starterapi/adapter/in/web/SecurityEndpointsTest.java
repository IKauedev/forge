package com.example.starterapi.adapter.in.web;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.containsInAnyOrder;
import static org.hamcrest.Matchers.containsString;
import static org.hamcrest.Matchers.equalTo;

import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.in.LookupAddressUseCase;
import com.example.starterapi.shared.security.Roles;
import io.quarkus.test.InjectMock;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.security.TestSecurity;
import io.quarkus.test.security.jwt.Claim;
import io.quarkus.test.security.jwt.JwtSecurity;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

/**
 * Garante que as regras de segurança realmente valem: esquecer uma anotação (ou usar uma que o
 * Quarkus ignora) deixa o endpoint aberto sem nenhum erro de compilação, e estes testes acusam.
 */
@QuarkusTest
class SecurityEndpointsTest {

  @InjectMock LookupAddressUseCase lookup;

  @Test
  void addressesRequireAuthentication() {
    given()
        .get("/api/addresses/01001000")
        .then()
        .statusCode(401)
        .contentType(containsString("application/problem+json"))
        .header("WWW-Authenticate", "Bearer")
        .body("title", equalTo("Não autenticado"))
        .body("status", equalTo(401));
  }

  @Test
  void anInvalidTokenIsAProblemDetailWith401() {
    given()
        .auth()
        .oauth2("isto.nao.e-um-jwt")
        .get("/api/addresses/01001000")
        .then()
        .statusCode(401)
        .contentType(containsString("application/problem+json"))
        .body("status", equalTo(401));
  }

  @Test
  @TestSecurity(user = "ana", roles = Roles.USER)
  void authenticatedUserCanLookUpAddresses() {
    Mockito.when(lookup.lookup("01001000"))
        .thenReturn(new Address("01001000", "Praça da Sé", "", "Sé", "São Paulo", "SP"));

    given().get("/api/addresses/01001000").then().statusCode(200);
  }

  @Test
  void adminAreaRequiresAuthentication() {
    given().get("/api/admin/ping").then().statusCode(401);
  }

  @Test
  @TestSecurity(user = "ana", roles = Roles.USER)
  void adminAreaIsForbiddenForCommonUsers() {
    given()
        .get("/api/admin/ping")
        .then()
        .statusCode(403)
        .contentType(containsString("application/problem+json"))
        .body("title", equalTo("Acesso negado"))
        .body("status", equalTo(403));
  }

  @Test
  @TestSecurity(user = "root", roles = Roles.ADMIN)
  void adminAreaIsOpenToAdmins() {
    given().get("/api/admin/ping").then().statusCode(200).body("status", equalTo("ok"));
  }

  @Test
  void meRequiresAuthentication() {
    given().get("/api/me").then().statusCode(401);
  }

  @Test
  @TestSecurity(
      user = "ana",
      roles = {Roles.USER, Roles.ADMIN})
  @JwtSecurity(
      claims = {
        @Claim(key = "preferred_username", value = "ana.souza"),
        @Claim(key = "scope", value = "orders:read orders:write")
      })
  void meDescribesTheAuthenticatedUser() {
    given()
        .get("/api/me")
        .then()
        .statusCode(200)
        .body("username", equalTo("ana.souza"))
        .body("roles", containsInAnyOrder("ADMIN", "USER"))
        .body("scopes", containsInAnyOrder("orders:read", "orders:write"));
  }

  @Test
  void healthIsPublic() {
    given().get("/q/health/ready").then().statusCode(200);
  }
}
