package com.example.starterapi.adapter.in.web;

import com.fasterxml.jackson.annotation.JsonInclude;
import org.jboss.resteasy.reactive.RestResponse;

/** Corpo de erro no formato RFC 9457 (application/problem+json). */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record ProblemDetail(String type, String title, int status, String detail) {

  public static final String MEDIA_TYPE = "application/problem+json";

  public static ProblemDetail of(int status, String title, String detail) {
    return new ProblemDetail("about:blank", title, status, detail);
  }

  /** Resposta application/problem+json com o status informado. */
  public static RestResponse<ProblemDetail> response(int status, String title, String detail) {
    return RestResponse.ResponseBuilder.<ProblemDetail>create(
            RestResponse.Status.fromStatusCode(status))
        .entity(of(status, title, detail))
        .type(MEDIA_TYPE)
        .build();
  }
}
