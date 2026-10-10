package com.example.starterapi.adapter.out.persistence;

import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.out.AddressRepositoryPort;
import java.util.Optional;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

@Component
public class AddressPersistenceAdapter implements AddressRepositoryPort {

  private final SpringDataAddressRepository repository;

  public AddressPersistenceAdapter(SpringDataAddressRepository repository) {
    this.repository = repository;
  }

  @Override
  @Transactional(readOnly = true)
  public Optional<Address> findByCep(String cep) {
    return repository.findById(cep).map(AddressPersistenceAdapter::toDomain);
  }

  @Override
  @Transactional
  public void save(Address a) {
    repository.save(
        new AddressEntity(
            a.cep(), a.street(), a.complement(), a.neighborhood(), a.city(), a.state()));
  }

  private static Address toDomain(AddressEntity e) {
    return new Address(
        e.getCep(),
        e.getStreet(),
        e.getComplement(),
        e.getNeighborhood(),
        e.getCity(),
        e.getState());
  }
}
