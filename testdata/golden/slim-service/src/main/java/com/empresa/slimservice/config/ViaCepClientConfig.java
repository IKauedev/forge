package com.empresa.slimservice.config;

import com.empresa.slimservice.adapter.out.client.ViaCepFeignAdapter;
import com.empresa.slimservice.adapter.out.client.ViaCepFeignClient;
import com.empresa.slimservice.adapter.out.client.ViaCepMapper;
import com.empresa.slimservice.domain.port.out.AddressLookupPort;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.cloud.openfeign.FeignClientBuilder;
import org.springframework.context.ApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

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
}
