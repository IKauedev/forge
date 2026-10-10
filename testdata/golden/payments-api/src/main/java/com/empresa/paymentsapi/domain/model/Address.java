package com.empresa.paymentsapi.domain.model;

public record Address(
    String cep, String street, String complement, String neighborhood, String city, String state) {}
