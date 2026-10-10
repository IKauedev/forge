package com.example.starterapi.shared.security;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.example.starterapi.config.MethodSecurityConfig;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.access.AccessDeniedException;
import org.springframework.security.test.context.support.WithAnonymousUser;
import org.springframework.security.test.context.support.WithMockUser;
import org.springframework.test.context.junit.jupiter.SpringJUnitConfig;

@SpringJUnitConfig({MethodSecurityConfig.class, SecurityAnnotationsTest.Config.class})
class SecurityAnnotationsTest {

  @Configuration(proxyBeanMethods = false)
  static class Config {
    @Bean
    Probe probe() {
      return new Probe();
    }
  }

  static class Probe {
    @IsAuthenticated
    public String authenticated() {
      return "ok";
    }

    @IsAdmin
    public String admin() {
      return "ok";
    }

    @HasRole("MANAGER")
    public String manager() {
      return "ok";
    }

    @HasAnyRole(roles = {"'USER'", "'ADMIN'"})
    public String userOrAdmin() {
      return "ok";
    }

    @HasScope("orders:read")
    public String ordersRead() {
      return "ok";
    }
  }

  @Autowired Probe probe;

  @Test
  @WithMockUser(roles = "USER")
  void isAdminDeniesUser() {
    assertThatThrownBy(probe::admin).isInstanceOf(AccessDeniedException.class);
  }

  @Test
  @WithMockUser(roles = "ADMIN")
  void isAdminAllowsAdmin() {
    assertThat(probe.admin()).isEqualTo("ok");
  }

  @Test
  @WithAnonymousUser
  void isAuthenticatedDeniesAnonymous() {
    assertThatThrownBy(probe::authenticated).isInstanceOf(AccessDeniedException.class);
  }

  @Test
  @WithMockUser
  void isAuthenticatedAllowsLoggedUser() {
    assertThat(probe.authenticated()).isEqualTo("ok");
  }

  @Test
  @WithMockUser(roles = "MANAGER")
  void hasRoleUsesTemplateValue() {
    assertThat(probe.manager()).isEqualTo("ok");
  }

  @Test
  @WithMockUser(roles = "USER")
  void hasRoleDeniesOtherRole() {
    assertThatThrownBy(probe::manager).isInstanceOf(AccessDeniedException.class);
  }

  @Test
  @WithMockUser(roles = "ADMIN")
  void hasAnyRoleAllowsOneOfThem() {
    assertThat(probe.userOrAdmin()).isEqualTo("ok");
  }

  @Test
  @WithMockUser(roles = "GUEST")
  void hasAnyRoleDeniesOthers() {
    assertThatThrownBy(probe::userOrAdmin).isInstanceOf(AccessDeniedException.class);
  }

  @Test
  @WithMockUser(authorities = "SCOPE_orders:read")
  void hasScopeChecksScopeAuthority() {
    assertThat(probe.ordersRead()).isEqualTo("ok");
  }

  @Test
  @WithMockUser(roles = "USER")
  void hasScopeDeniesWithoutScope() {
    assertThatThrownBy(probe::ordersRead).isInstanceOf(AccessDeniedException.class);
  }
}
