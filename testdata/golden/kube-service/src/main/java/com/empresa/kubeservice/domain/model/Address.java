package com.empresa.kubeservice.domain.model;

import java.io.Serializable;

/** Serializable permite guardar o endereço em cache distribuído (Redis). */
public record Address(
    String cep, String street, String complement, String neighborhood, String city, String state)
    implements Serializable {}
