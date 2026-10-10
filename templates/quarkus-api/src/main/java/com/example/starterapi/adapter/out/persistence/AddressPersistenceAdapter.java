package com.example.starterapi.adapter.out.persistence;

import com.example.starterapi.domain.model.Address;
import com.example.starterapi.domain.port.out.AddressRepositoryPort;
import io.quarkus.hibernate.orm.panache.PanacheRepositoryBase;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.transaction.Transactional;
import java.util.Optional;

@ApplicationScoped
public class AddressPersistenceAdapter
    implements AddressRepositoryPort, PanacheRepositoryBase<AddressEntity, String> {

  @Override
  @Transactional
  public Optional<Address> findByCep(String cep) {
    return findByIdOptional(cep).map(AddressPersistenceAdapter::toDomain);
  }

  @Override
  @Transactional
  public void save(Address a) {
    // merge: o endereço pode já existir se duas requisições consultarem o mesmo CEP ao mesmo tempo
    getEntityManager()
        .merge(
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
