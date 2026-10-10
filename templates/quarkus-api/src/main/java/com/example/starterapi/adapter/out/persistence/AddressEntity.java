package com.example.starterapi.adapter.out.persistence;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

@Entity
@Table(name = "addresses")
public class AddressEntity {

  @Id
  @Column(length = 8)
  private String cep;

  private String street;
  private String complement;
  private String neighborhood;
  private String city;

  @Column(length = 2)
  private String state;

  protected AddressEntity() {}

  public AddressEntity(
      String cep,
      String street,
      String complement,
      String neighborhood,
      String city,
      String state) {
    this.cep = cep;
    this.street = street;
    this.complement = complement;
    this.neighborhood = neighborhood;
    this.city = city;
    this.state = state;
  }

  public String getCep() {
    return cep;
  }

  public String getStreet() {
    return street;
  }

  public String getComplement() {
    return complement;
  }

  public String getNeighborhood() {
    return neighborhood;
  }

  public String getCity() {
    return city;
  }

  public String getState() {
    return state;
  }
}
