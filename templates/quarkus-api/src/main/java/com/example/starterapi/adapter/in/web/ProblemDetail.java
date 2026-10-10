package com.example.starterapi.adapter.in.web;

import com.fasterxml.jackson.annotation.JsonInclude;

/** Corpo de erro no formato RFC 9457 (application/problem+json). */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record ProblemDetail(String type, String title, int status, String detail) {

  public static final String MEDIA_TYPE = "application/problem+json";

  public static ProblemDetail of(int status, String title, String detail) {
    return new ProblemDetail("about:blank", title, status, detail);
  }
}
