package com.empresa.kubeservice.adapter.in.web;

import com.empresa.kubeservice.domain.model.Address;

public record AddressResponse(
    String cep, String street, String complement, String neighborhood, String city, String state) {

  static AddressResponse from(Address a) {
    return new AddressResponse(
        a.cep(), a.street(), a.complement(), a.neighborhood(), a.city(), a.state());
  }
}
