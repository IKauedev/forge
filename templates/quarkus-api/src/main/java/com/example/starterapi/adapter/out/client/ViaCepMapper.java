package com.example.starterapi.adapter.out.client;

import com.example.starterapi.domain.model.Address;

final class ViaCepMapper {

  private ViaCepMapper() {}

  static Address toDomain(ViaCepResponse r) {
    return new Address(
        r.cep() == null ? null : r.cep().replaceAll("\\D", ""),
        r.logradouro(),
        r.complemento(),
        r.bairro(),
        r.localidade(),
        r.uf());
  }
}
