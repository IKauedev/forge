package com.empresa.paymentsapi.shared;

public final class CorrelationId {

  public static final String HEADER = "X-Correlation-Id";
  public static final String MDC_KEY = "correlationId";

  private CorrelationId() {}
}
