package com.empresa.ordersservice.shared.security;

import java.lang.annotation.Documented;
import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import org.springframework.security.access.prepost.PreAuthorize;

/** Exige a role informada, ex.: {@code @HasRole("MANAGER")}. */
@Target({ElementType.METHOD, ElementType.TYPE})
@Retention(RetentionPolicy.RUNTIME)
@Documented
@PreAuthorize("hasRole('{value}')")
public @interface HasRole {
  /** Nome da role, sem o prefixo ROLE_. */
  String value();
}
