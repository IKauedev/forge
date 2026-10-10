package com.empresa.paymentsapi.adapter.in.web;

import static org.mockito.BDDMockito.given;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.empresa.paymentsapi.config.SecurityConfig;
import com.empresa.paymentsapi.domain.exception.AddressNotFoundException;
import com.empresa.paymentsapi.domain.exception.ExternalServiceException;
import com.empresa.paymentsapi.domain.exception.InvalidCepException;
import com.empresa.paymentsapi.domain.model.Address;
import com.empresa.paymentsapi.domain.port.in.LookupAddressUseCase;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.webmvc.test.autoconfigure.WebMvcTest;
import org.springframework.context.annotation.Import;
import org.springframework.http.MediaType;
import org.springframework.security.oauth2.jwt.JwtDecoder;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.request.MockHttpServletRequestBuilder;

@WebMvcTest(AddressController.class)
@Import({GlobalExceptionHandler.class, SecurityConfig.class})
class AddressControllerTest {
  @Autowired MockMvc mvc;
  @MockitoBean LookupAddressUseCase useCase;

  @MockitoBean JwtDecoder jwtDecoder;

  @Test
  void returnsAddress() throws Exception {
    given(useCase.lookup("01001000"))
        .willReturn(new Address("01001000", "Praça da Sé", "", "Sé", "São Paulo", "SP"));

    mvc.perform(authed(get("/api/addresses/01001000")))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.city").value("São Paulo"))
        .andExpect(jsonPath("$.street").value("Praça da Sé"));
  }

  @Test
  void invalidCepIsProblemDetail400() throws Exception {
    given(useCase.lookup("abc")).willThrow(new InvalidCepException("abc"));

    mvc.perform(authed(get("/api/addresses/abc")))
        .andExpect(status().isBadRequest())
        .andExpect(content().contentTypeCompatibleWith(MediaType.APPLICATION_PROBLEM_JSON))
        .andExpect(jsonPath("$.title").value("CEP inválido"));
  }

  @Test
  void notFoundIsProblemDetail404() throws Exception {
    given(useCase.lookup("99999999")).willThrow(new AddressNotFoundException("99999999"));

    mvc.perform(authed(get("/api/addresses/99999999"))).andExpect(status().isNotFound());
  }

  @Test
  void externalFailureIs503() throws Exception {
    given(useCase.lookup("01001000"))
        .willThrow(new ExternalServiceException("falha", new RuntimeException()));

    mvc.perform(authed(get("/api/addresses/01001000"))).andExpect(status().isServiceUnavailable());
  }

  private static MockHttpServletRequestBuilder authed(MockHttpServletRequestBuilder request) {
    return request.with(jwt());
  }
}
