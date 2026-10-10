package com.example.starterapi.adapter.in.web;

import static org.mockito.BDDMockito.given;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.example.starterapi.config.SecurityConfig;
import com.example.starterapi.domain.exception.AddressNotFoundException;
import com.example.starterapi.domain.exception.ExternalServiceException;
import com.example.starterapi.domain.exception.InvalidCepException;
import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.in.LookupAddressUseCase;
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
// forge:if auth!=none
@Import({GlobalExceptionHandler.class, SecurityConfig.class})
// forge:else
// ~ @Import(GlobalExceptionHandler.class)
// forge:end
class AddressControllerTest {

  @Autowired MockMvc mvc;
  @MockitoBean LookupAddressUseCase useCase;

  // forge:if auth!=none
  @MockitoBean JwtDecoder jwtDecoder;

  // forge:end
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

  // forge:if auth!=none
  private static MockHttpServletRequestBuilder authed(MockHttpServletRequestBuilder request) {
    return request.with(jwt());
  }

  // forge:else
  // ~ private static MockHttpServletRequestBuilder authed(MockHttpServletRequestBuilder request) {
  // ~   return request;
  // ~ }
  // forge:end
}
