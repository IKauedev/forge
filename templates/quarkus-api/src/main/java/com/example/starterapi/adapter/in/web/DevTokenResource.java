package com.example.starterapi.adapter.in.web;

import io.quarkus.arc.profile.IfBuildProfile;
import io.smallrye.jwt.build.Jwt;
import jakarta.annotation.security.PermitAll;
import jakarta.ws.rs.DefaultValue;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.QueryParam;
import jakarta.ws.rs.core.MediaType;
import java.time.Duration;
import java.util.Arrays;
import java.util.Map;
import java.util.Set;
import java.util.TreeSet;
import org.eclipse.microprofile.config.inject.ConfigProperty;

/**
 * Emite tokens de teste. Existe SOMENTE nos perfis dev e test: o {@code @IfBuildProfile} tira o
 * endpoint do build de produção.
 */
@Path("/dev")
@IfBuildProfile(anyOf = {"dev", "test"})
@PermitAll
@Produces(MediaType.APPLICATION_JSON)
public class DevTokenResource {

  private static final Duration TTL = Duration.ofHours(1);

  private final String secret;
  private final String issuer;
  private final String rolesClaim;

  public DevTokenResource(
      @ConfigProperty(name = "app.security.dev-secret") String secret,
      @ConfigProperty(name = "mp.jwt.verify.issuer") String issuer,
      @ConfigProperty(name = "smallrye.jwt.path.groups") String rolesClaim) {
    this.secret = secret;
    this.issuer = issuer;
    this.rolesClaim = rolesClaim;
  }

  /** Ex.: {@code POST /dev/token?sub=ana&roles=ADMIN,USER&scopes=orders:read}. */
  @POST
  @Path("/token")
  public Map<String, Object> token(
      @QueryParam("sub") @DefaultValue("dev-user") String sub,
      @QueryParam("roles") @DefaultValue("USER") String roles,
      @QueryParam("scopes") @DefaultValue("") String scopes) {
    String token =
        Jwt.issuer(issuer)
            .subject(sub)
            .expiresIn(TTL)
            .claim(rolesClaim, split(roles))
            .claim("scope", String.join(" ", split(scopes)))
            .signWithSecret(secret);
    return Map.of("access_token", token, "token_type", "Bearer", "expires_in", TTL.toSeconds());
  }

  private static Set<String> split(String csv) {
    return new TreeSet<>(Arrays.stream(csv.split("[,\\s]+")).filter(s -> !s.isBlank()).toList());
  }
}
