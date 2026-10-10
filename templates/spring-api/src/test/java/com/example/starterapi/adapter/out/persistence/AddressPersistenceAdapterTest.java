package com.example.starterapi.adapter.out.persistence;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.starterapi.TestcontainersConfig;
import com.example.starterapi.domain.model.Address;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.data.jpa.test.autoconfigure.DataJpaTest;
import org.springframework.boot.jdbc.test.autoconfigure.AutoConfigureTestDatabase;
import org.springframework.context.annotation.Import;

@DataJpaTest
@AutoConfigureTestDatabase(replace = AutoConfigureTestDatabase.Replace.NONE)
@Import({TestcontainersConfig.class, AddressPersistenceAdapter.class})
class AddressPersistenceAdapterTest {

  @Autowired AddressPersistenceAdapter adapter;

  @Test
  void savesAndLoadsThroughFlywayMigratedSchema() {
    Address address = new Address("01001000", "Praça da Sé", "", "Sé", "São Paulo", "SP");

    adapter.save(address);

    assertThat(adapter.findByCep("01001000")).contains(address);
    assertThat(adapter.findByCep("00000000")).isEmpty();
  }
}
