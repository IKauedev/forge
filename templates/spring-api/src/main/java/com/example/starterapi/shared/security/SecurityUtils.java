package com.example.starterapi.shared.security;

import java.util.Optional;
import java.util.Set;
import java.util.TreeSet;
import org.springframework.security.authentication.AnonymousAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.oauth2.jwt.Jwt;

/** Acesso ao usuário autenticado fora dos controllers (adapters, filtros, auditoria). */
public final class SecurityUtils {

  private static final String ROLE_PREFIX = "ROLE_";
  private static final String SCOPE_PREFIX = "SCOPE_";

  private SecurityUtils() {}

  public static Optional<CurrentUser> currentUser() {
    Authentication auth = SecurityContextHolder.getContext().getAuthentication();
    if (auth == null || !auth.isAuthenticated() || auth instanceof AnonymousAuthenticationToken) {
      return Optional.empty();
    }
    return Optional.of(
        new CurrentUser(
            auth.getName(),
            username(auth),
            authorities(auth, ROLE_PREFIX),
            authorities(auth, SCOPE_PREFIX)));
  }

  public static boolean hasRole(String role) {
    return currentUser().map(u -> u.hasRole(role)).orElse(false);
  }

  private static String username(Authentication auth) {
    if (auth.getPrincipal() instanceof Jwt jwt) {
      for (String claim : new String[] {"preferred_username", "cognito:username", "email"}) {
        String value = jwt.getClaimAsString(claim);
        if (value != null && !value.isBlank()) {
          return value;
        }
      }
    }
    return auth.getName();
  }

  private static Set<String> authorities(Authentication auth, String prefix) {
    Set<String> result = new TreeSet<>();
    for (GrantedAuthority authority : auth.getAuthorities()) {
      String name = authority.getAuthority();
      if (name != null && name.startsWith(prefix)) {
        result.add(name.substring(prefix.length()));
      }
    }
    return Set.copyOf(result);
  }
}
