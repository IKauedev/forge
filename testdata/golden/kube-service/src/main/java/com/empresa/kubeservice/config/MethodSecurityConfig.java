package com.empresa.kubeservice.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;
import org.springframework.security.core.annotation.AnnotationTemplateExpressionDefaults;

/** Habilita @PreAuthorize e as meta-anotações com parâmetros de shared.security. */
@Configuration(proxyBeanMethods = false)
@EnableMethodSecurity
public class MethodSecurityConfig {

  /** Necessário para os placeholders ({value}, {roles}) das anotações personalizadas. */
  @Bean
  static AnnotationTemplateExpressionDefaults templateExpressionDefaults() {
    return new AnnotationTemplateExpressionDefaults();
  }
}
