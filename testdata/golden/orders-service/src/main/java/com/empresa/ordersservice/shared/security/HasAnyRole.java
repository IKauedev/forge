package com.empresa.ordersservice.shared.security;

import java.lang.annotation.Documented;
import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import org.springframework.security.access.prepost.PreAuthorize;

/** Exige qualquer uma das roles: {@code @HasAnyRole(roles = {"'USER'", "'ADMIN'"})}. */
@Target({ElementType.METHOD, ElementType.TYPE})
@Retention(RetentionPolicy.RUNTIME)
@Documented
@PreAuthorize("hasAnyRole({roles})")
public @interface HasAnyRole {
  /**
   * Roles entre aspas simples, pois entram direto na expressão SpEL: {@code {"'USER'", "'ADMIN'"}}.
   */
  String[] roles();
}
