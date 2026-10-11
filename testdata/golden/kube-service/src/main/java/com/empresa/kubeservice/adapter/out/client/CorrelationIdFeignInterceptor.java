package com.empresa.kubeservice.adapter.out.client;

import com.empresa.kubeservice.shared.CorrelationId;
import feign.RequestInterceptor;
import feign.RequestTemplate;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

@Component
public class CorrelationIdFeignInterceptor implements RequestInterceptor {

  @Override
  public void apply(RequestTemplate template) {
    String id = MDC.get(CorrelationId.MDC_KEY);
    if (id != null) {
      template.header(CorrelationId.HEADER, id);
    }
  }
}
