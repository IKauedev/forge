package com.empresa.paymentsapi.config;

import com.nimbusds.jose.jwk.source.ImmutableSecret;
import java.nio.charset.StandardCharsets;
import javax.crypto.SecretKey;
import javax.crypto.spec.SecretKeySpec;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;
import org.springframework.security.oauth2.jose.jws.MacAlgorithm;
import org.springframework.security.oauth2.jwt.JwtDecoder;
import org.springframework.security.oauth2.jwt.JwtEncoder;
import org.springframework.security.oauth2.jwt.NimbusJwtDecoder;
import org.springframework.security.oauth2.jwt.NimbusJwtEncoder;

/**
 * JWT HS256 com segredo compartilhado, apenas para desenvolvimento local e testes. Nos perfis dev e
 * prod o decoder vem de spring.security.oauth2.resourceserver.jwt.issuer-uri.
 */
@Configuration(proxyBeanMethods = false)
@Profile({"local", "test"})
class DevJwtConfig {

  @Bean
  SecretKey devJwtKey(@Value("${app.security.dev-secret}") String secret) {
    return new SecretKeySpec(secret.getBytes(StandardCharsets.UTF_8), "HmacSHA256");
  }

  @Bean
  JwtDecoder jwtDecoder(SecretKey devJwtKey) {
    return NimbusJwtDecoder.withSecretKey(devJwtKey).macAlgorithm(MacAlgorithm.HS256).build();
  }

  @Bean
  JwtEncoder jwtEncoder(SecretKey devJwtKey) {
    return new NimbusJwtEncoder(new ImmutableSecret<>(devJwtKey));
  }
}
