package com.example.starterapi.shared.security;

import io.quarkus.security.identity.SecurityIdentity;
import java.util.Arrays;
import java.util.Set;
import java.util.TreeSet;
import org.eclipse.microprofile.jwt.JsonWebToken;

/** Visão imutável do usuário autenticado na requisição atual. */
public record CurrentUser(String subject, String username, Set<String> roles, Set<String> scopes) {

  public static CurrentUser from(SecurityIdentity identity) {
    String subject = identity.getPrincipal().getName();
    String username = subject;
    Set<String> scopes = Set.of();
    if (identity.getPrincipal() instanceof JsonWebToken jwt) {
      username = firstNonBlank(jwt, subject, "preferred_username", "cognito:username", "email");
      Object scope = jwt.getClaim("scope");
      if (scope instanceof String s && !s.isBlank()) {
        scopes = new TreeSet<>(Arrays.asList(s.trim().split("\\s+")));
      }
    }
    return new CurrentUser(
        subject, username, Set.copyOf(new TreeSet<>(identity.getRoles())), scopes);
  }

  public boolean hasRole(String role) {
    return roles.contains(role);
  }

  public boolean hasScope(String scope) {
    return scopes.contains(scope);
  }

  public boolean isAdmin() {
    return hasRole(Roles.ADMIN);
  }

  private static String firstNonBlank(JsonWebToken jwt, String fallback, String... claims) {
    for (String claim : claims) {
      Object value = jwt.getClaim(claim);
      if (value instanceof String s && !s.isBlank()) {
        return s;
      }
    }
    return fallback;
  }
}
