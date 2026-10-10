package com.example.starterapi.adapter.in.web;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.containsInAnyOrder;
import static org.hamcrest.Matchers.equalTo;

import io.quarkus.test.junit.QuarkusTest;
import org.junit.jupiter.api.Test;

/** Fluxo real, sem @TestSecurity: emite um token em /dev/token e o backend o verifica (HS256). */
@QuarkusTest
class DevTokenResourceTest {

  private static String token(String query) {
    return given()
        .post("/dev/token" + query)
        .then()
        .statusCode(200)
        .body("token_type", equalTo("Bearer"))
        .extract()
        .path("access_token");
  }

  @Test
  void issuedTokenIsAcceptedByTheApi() {
    String token = token("?sub=ana&roles=ADMIN,USER&scopes=orders:read");

    given()
        .auth()
        .oauth2(token)
        .get("/api/me")
        .then()
        .statusCode(200)
        .body("subject", equalTo("ana"))
        .body("roles", containsInAnyOrder("ADMIN", "USER"))
        .body("scopes", containsInAnyOrder("orders:read"));
    given().auth().oauth2(token).get("/api/admin/ping").then().statusCode(200);
  }

  @Test
  void tokenWithoutAdminRoleIsForbiddenInTheAdminArea() {
    String token = token("?roles=USER");

    given().auth().oauth2(token).get("/api/admin/ping").then().statusCode(403);
  }

  @Test
  void tamperedTokenIsRejected() {
    String token = token("");

    given().auth().oauth2(token + "x").get("/api/me").then().statusCode(401);
  }
}
