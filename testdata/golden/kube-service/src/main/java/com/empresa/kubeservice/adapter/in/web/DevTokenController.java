package com.empresa.kubeservice.adapter.in.web;

import java.time.Duration;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Profile;
import org.springframework.security.oauth2.jose.jws.MacAlgorithm;
import org.springframework.security.oauth2.jwt.JwsHeader;
import org.springframework.security.oauth2.jwt.JwtClaimsSet;
import org.springframework.security.oauth2.jwt.JwtEncoder;
import org.springframework.security.oauth2.jwt.JwtEncoderParameters;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

/** Emite tokens de teste. Existe SOMENTE no perfil local. */
@RestController
@Profile("local")
@RequestMapping("/dev")
public class DevTokenController {

  private static final Duration TTL = Duration.ofHours(1);

  private final JwtEncoder encoder;
  private final String rolesClaim;

  public DevTokenController(
      JwtEncoder encoder, @Value("${app.security.roles-claim:roles}") String rolesClaim) {
    this.encoder = encoder;
    this.rolesClaim = rolesClaim;
  }

  /** Ex.: {@code POST /dev/token?sub=ana&roles=ADMIN,USER&scopes=orders:read}. */
  @PostMapping("/token")
  public Map<String, Object> token(
      @RequestParam(defaultValue = "dev-user") String sub,
      @RequestParam(defaultValue = "USER") List<String> roles,
      @RequestParam(defaultValue = "") List<String> scopes) {
    Instant now = Instant.now();
    JwtClaimsSet claims =
        JwtClaimsSet.builder()
            .issuer("kube-service-dev")
            .subject(sub)
            .issuedAt(now)
            .expiresAt(now.plus(TTL))
            .claim(rolesClaim, roles)
            .claim("scope", String.join(" ", scopes))
            .build();
    String token =
        encoder
            .encode(JwtEncoderParameters.from(JwsHeader.with(MacAlgorithm.HS256).build(), claims))
            .getTokenValue();
    return Map.of("access_token", token, "token_type", "Bearer", "expires_in", TTL.toSeconds());
  }
}
