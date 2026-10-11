package com.empresa.kubeservice.domain.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.empresa.kubeservice.domain.exception.AddressNotFoundException;
import com.empresa.kubeservice.domain.exception.InvalidCepException;
import com.empresa.kubeservice.domain.model.Address;
import com.empresa.kubeservice.domain.port.out.AddressLookupPort;
import com.empresa.kubeservice.domain.port.out.AddressRepositoryPort;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicInteger;
import org.junit.jupiter.api.Test;

class AddressLookupServiceTest {

  private static final Address SE =
      new Address("01001000", "Praça da Sé", "lado ímpar", "Sé", "São Paulo", "SP");

  private final Map<String, Address> stored = new HashMap<>();
  private final AtomicInteger remoteCalls = new AtomicInteger();
  private Optional<Address> remoteResult = Optional.of(SE);

  private final AddressLookupPort remote =
      cep -> {
        remoteCalls.incrementAndGet();
        return remoteResult;
      };

  private final AddressRepositoryPort repository =
      new AddressRepositoryPort() {
        @Override
        public Optional<Address> findByCep(String cep) {
          return Optional.ofNullable(stored.get(cep));
        }

        @Override
        public void save(Address address) {
          stored.put(address.cep(), address);
        }
      };

  private final AddressLookupService service = new AddressLookupService(remote, repository);

  @Test
  void normalizesCepAndFetchesFromRemote() {
    assertThat(service.lookup("01001-000")).isEqualTo(SE);
    assertThat(remoteCalls).hasValue(1);
  }

  @Test
  void secondLookupIsServedFromRepository() {
    service.lookup("01001000");
    service.lookup("01001000");
    assertThat(remoteCalls).hasValue(1);
  }

  @Test
  void rejectsInvalidCep() {
    assertThatThrownBy(() -> service.lookup("123")).isInstanceOf(InvalidCepException.class);
    assertThatThrownBy(() -> service.lookup(null)).isInstanceOf(InvalidCepException.class);
  }

  @Test
  void throwsWhenNotFoundAnywhere() {
    remoteResult = Optional.empty();
    assertThatThrownBy(() -> service.lookup("99999999"))
        .isInstanceOf(AddressNotFoundException.class);
  }
}
