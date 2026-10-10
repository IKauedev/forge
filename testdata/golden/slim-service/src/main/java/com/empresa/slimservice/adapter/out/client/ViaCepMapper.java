package com.empresa.slimservice.adapter.out.client;

import com.empresa.slimservice.domain.model.Address;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.Named;

@Mapper(componentModel = "spring")
public interface ViaCepMapper {

  @Mapping(target = "cep", source = "cep", qualifiedByName = "digitsOnly")
  @Mapping(target = "street", source = "logradouro")
  @Mapping(target = "complement", source = "complemento")
  @Mapping(target = "neighborhood", source = "bairro")
  @Mapping(target = "city", source = "localidade")
  @Mapping(target = "state", source = "uf")
  Address toDomain(ViaCepResponse response);

  @Named("digitsOnly")
  default String digitsOnly(String cep) {
    return cep == null ? null : cep.replaceAll("\\D", "");
  }
}
