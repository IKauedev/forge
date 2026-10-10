package com.empresa.ordersservice.adapter.in.web;

import static org.assertj.core.api.Assertions.assertThat;

import com.nimbusds.jose.jwk.source.ImmutableSecret;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import javax.crypto.SecretKey;
import javax.crypto.spec.SecretKeySpec;
import org.junit.jupiter.api.Test;
import org.springframework.security.oauth2.jose.jws.MacAlgorithm;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.security.oauth2.jwt.NimbusJwtDecoder;
import org.springframework.security.oauth2.jwt.NimbusJwtEncoder;

class DevTokenControllerTest {

  private static final SecretKey KEY =
      new SecretKeySpec(
          "test-only-secret-0123456789-0123456789".getBytes(StandardCharsets.UTF_8), "HmacSHA256");

  @Test
  void issuesTokenThatTheDevDecoderAccepts() {
    var controller =
        new DevTokenController(new NimbusJwtEncoder(new ImmutableSecret<>(KEY)), "roles");

    Map<String, Object> response =
        controller.token("ana", List.of("ADMIN", "USER"), List.of("orders:read"));

    Jwt jwt =
        NimbusJwtDecoder.withSecretKey(KEY)
            .macAlgorithm(MacAlgorithm.HS256)
            .build()
            .decode((String) response.get("access_token"));
    assertThat(jwt.getSubject()).isEqualTo("ana");
    assertThat(jwt.getClaimAsStringList("roles")).containsExactly("ADMIN", "USER");
    assertThat(jwt.getClaimAsString("scope")).isEqualTo("orders:read");
    assertThat(response).containsEntry("token_type", "Bearer");
  }
}
