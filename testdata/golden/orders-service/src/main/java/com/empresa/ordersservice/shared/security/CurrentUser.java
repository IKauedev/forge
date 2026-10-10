package com.empresa.ordersservice.shared.security;

import java.util.Set;

/** Visão imutável do usuário autenticado na requisição atual. */
public record CurrentUser(String subject, String username, Set<String> roles, Set<String> scopes) {

  public boolean hasRole(String role) {
    return roles.contains(role);
  }

  public boolean hasScope(String scope) {
    return scopes.contains(scope);
  }

  public boolean isAdmin() {
    return hasRole(Roles.ADMIN);
  }
}
