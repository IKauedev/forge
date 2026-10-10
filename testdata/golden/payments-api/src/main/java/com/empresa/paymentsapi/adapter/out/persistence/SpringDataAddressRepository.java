package com.empresa.paymentsapi.adapter.out.persistence;

import org.springframework.data.jpa.repository.JpaRepository;

public interface SpringDataAddressRepository extends JpaRepository<AddressEntity, String> {}
