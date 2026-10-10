package com.example.starterapi.adapter.out.persistence;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.out.AddressRepositoryPort;
import io.quarkus.test.junit.QuarkusTest;
import jakarta.inject.Inject;
import org.junit.jupiter.api.Test;

/** Usa o banco do Dev Services (Docker) com as migrations Flyway reais. */
@QuarkusTest
class AddressPersistenceAdapterTest {

  @Inject AddressRepositoryPort repository;

  @Test
  void savesAndFindsAnAddress() {
    Address a = new Address("01310100", "Avenida Paulista", "", "Bela Vista", "São Paulo", "SP");

    repository.save(a);

    assertThat(repository.findByCep("01310100")).contains(a);
  }

  @Test
  void savingTheSameCepTwiceKeepsOneRow() {
    repository.save(new Address("20040020", "Rua A", "", "Centro", "Rio de Janeiro", "RJ"));
    repository.save(new Address("20040020", "Rua B", "", "Centro", "Rio de Janeiro", "RJ"));

    assertThat(repository.findByCep("20040020")).map(Address::street).contains("Rua B");
  }

  @Test
  void unknownCepIsEmpty() {
    assertThat(repository.findByCep("00000001")).isEmpty();
  }
}
