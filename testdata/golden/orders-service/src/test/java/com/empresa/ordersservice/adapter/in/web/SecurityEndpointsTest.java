package com.empresa.ordersservice.adapter.in.web;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.empresa.ordersservice.config.SecurityConfig;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.webmvc.test.autoconfigure.WebMvcTest;
import org.springframework.context.annotation.Import;
import org.springframework.http.MediaType;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.oauth2.jwt.JwtDecoder;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.test.web.servlet.MockMvc;

@WebMvcTest({MeController.class, AdminController.class})
@Import(SecurityConfig.class)
class SecurityEndpointsTest {

  @Autowired MockMvc mvc;
  @MockitoBean JwtDecoder jwtDecoder;

  @Test
  void withoutTokenIs401ProblemDetail() throws Exception {
    mvc.perform(get("/api/me"))
        .andExpect(status().isUnauthorized())
        .andExpect(content().contentTypeCompatibleWith(MediaType.APPLICATION_PROBLEM_JSON))
        .andExpect(jsonPath("$.status").value(401));
  }

  @Test
  void currentUserExposesRolesAndScopes() throws Exception {
    mvc.perform(
            get("/api/me")
                .with(
                    jwt()
                        .jwt(j -> j.subject("u-1").claim("preferred_username", "ana"))
                        .authorities(
                            new SimpleGrantedAuthority("ROLE_USER"),
                            new SimpleGrantedAuthority("SCOPE_orders:read"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.subject").value("u-1"))
        .andExpect(jsonPath("$.username").value("ana"))
        .andExpect(jsonPath("$.roles[0]").value("USER"))
        .andExpect(jsonPath("$.scopes[0]").value("orders:read"));
  }

  @Test
  void userCannotReachAdminEndpoint() throws Exception {
    mvc.perform(
            get("/api/admin/ping").with(jwt().authorities(new SimpleGrantedAuthority("ROLE_USER"))))
        .andExpect(status().isForbidden())
        .andExpect(content().contentTypeCompatibleWith(MediaType.APPLICATION_PROBLEM_JSON))
        .andExpect(jsonPath("$.status").value(403));
  }

  @Test
  void adminReachesAdminEndpoint() throws Exception {
    mvc.perform(
            get("/api/admin/ping")
                .with(jwt().authorities(new SimpleGrantedAuthority("ROLE_ADMIN"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.status").value("ok"));
  }
}
