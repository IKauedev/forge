package com.empresa.paymentsapi.config;

import com.empresa.paymentsapi.adapter.out.client.ViaCepFeignAdapter;
import com.empresa.paymentsapi.adapter.out.client.ViaCepFeignClient;
import com.empresa.paymentsapi.adapter.out.client.ViaCepMapper;
import com.empresa.paymentsapi.adapter.out.client.ViaCepWebClientAdapter;
import com.empresa.paymentsapi.domain.port.out.AddressLookupPort;
import com.empresa.paymentsapi.shared.CorrelationId;
import java.time.Duration;
import org.slf4j.MDC;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.cloud.openfeign.FeignClientBuilder;
import org.springframework.context.ApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.reactive.function.client.ExchangeFilterFunction;
import org.springframework.web.reactive.function.client.WebClient;

/** Escolhe a implementação do cliente ViaCEP por clients.viacep.impl (feign | webclient). */
@Configuration
public class ViaCepClientConfig {

  @Configuration
  @ConditionalOnProperty(name = "clients.viacep.impl", havingValue = "feign", matchIfMissing = true)
  static class FeignImpl {

    @Bean
    ViaCepFeignClient viaCepFeignClient(
        ApplicationContext context, @Value("${clients.viacep.url}") String url) {
      return new FeignClientBuilder(context)
          .forType(ViaCepFeignClient.class, "viacep")
          .url(url)
          .build();
    }

    @Bean
    AddressLookupPort viaCepLookup(ViaCepFeignClient client, ViaCepMapper mapper) {
      return new ViaCepFeignAdapter(client, mapper);
    }
  }

  @Configuration
  @ConditionalOnProperty(name = "clients.viacep.impl", havingValue = "webclient")
  static class WebClientImpl {

    @Bean
    WebClient viaCepWebClient(
        WebClient.Builder builder, @Value("${clients.viacep.url}") String url) {
      return builder
          .baseUrl(url)
          .filter(
              ExchangeFilterFunction.ofRequestProcessor(
                  request -> {
                    String id = MDC.get(CorrelationId.MDC_KEY);
                    if (id == null) {
                      return reactor.core.publisher.Mono.just(request);
                    }
                    return reactor.core.publisher.Mono.just(
                        org.springframework.web.reactive.function.client.ClientRequest.from(request)
                            .header(CorrelationId.HEADER, id)
                            .build());
                  }))
          .build();
    }

    @Bean
    AddressLookupPort viaCepLookup(
        WebClient viaCepWebClient,
        ViaCepMapper mapper,
        @Value("${clients.viacep.read-timeout-ms:3000}") long timeoutMs) {
      return new ViaCepWebClientAdapter(viaCepWebClient, mapper, Duration.ofMillis(timeoutMs));
    }
  }
}
