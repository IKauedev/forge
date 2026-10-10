package com.example.starterapi.shared.security;

/**
 * Nomes de roles, para usar com {@code @RolesAllowed(Roles.ADMIN)}.
 *
 * <p>Use as anotações do Quarkus ({@code @Authenticated}, {@code @RolesAllowed}) diretamente: uma
 * anotação própria que apenas agrega {@code @RolesAllowed} (meta-anotação ou {@code @Stereotype}) é
 * ignorada pelo Quarkus e deixaria o endpoint aberto sem aviso.
 */
public final class Roles {

  public static final String ADMIN = "ADMIN";
  public static final String USER = "USER";

  private Roles() {}
}
