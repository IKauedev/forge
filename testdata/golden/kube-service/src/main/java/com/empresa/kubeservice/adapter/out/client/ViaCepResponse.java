package com.empresa.kubeservice.adapter.out.client;

/** Contrato externo do ViaCEP. Não vaza para o domínio. */
public record ViaCepResponse(
    String cep,
    String logradouro,
    String complemento,
    String bairro,
    String localidade,
    String uf,
    Boolean erro) {}
