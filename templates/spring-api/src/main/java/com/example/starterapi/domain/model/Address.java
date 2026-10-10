package com.example.starterapi.domain.model;

public record Address(
    String cep, String street, String complement, String neighborhood, String city, String state) {}
