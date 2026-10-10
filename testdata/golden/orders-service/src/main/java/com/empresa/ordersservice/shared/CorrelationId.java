package com.empresa.ordersservice.shared;

public final class CorrelationId {

  public static final String HEADER = "X-Correlation-Id";
  public static final String MDC_KEY = "correlationId";

  private CorrelationId() {}
}
