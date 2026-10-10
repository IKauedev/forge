package com.empresa.slimservice.adapter.out.client;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

/** Criado por {@code ViaCepClientConfig} apenas quando clients.viacep.impl=feign. */
public interface ViaCepFeignClient {

  @GetMapping("/{cep}/json")
  ViaCepResponse find(@PathVariable("cep") String cep);
}
