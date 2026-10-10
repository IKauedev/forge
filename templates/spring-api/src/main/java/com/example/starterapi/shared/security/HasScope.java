package com.example.starterapi.shared.security;

import java.lang.annotation.Documented;
import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import org.springframework.security.access.prepost.PreAuthorize;

/** Exige o scope OAuth2 informado, ex.: {@code @HasScope("orders:read")}. */
@Target({ElementType.METHOD, ElementType.TYPE})
@Retention(RetentionPolicy.RUNTIME)
@Documented
@PreAuthorize("hasAuthority('SCOPE_{value}')")
public @interface HasScope {
  /** Nome do scope OAuth2, sem o prefixo SCOPE_. */
  String value();
}
