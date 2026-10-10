package com.example.starterapi.adapter.in.web;

import com.example.starterapi.domain.port.in.LookupAddressUseCase;
import io.quarkus.security.Authenticated;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.PathParam;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;

@Path("/api/addresses")
@Produces(MediaType.APPLICATION_JSON)
// forge:if auth!=none
@Authenticated
// forge:end
public class AddressResource {

  private final LookupAddressUseCase lookupAddress;

  public AddressResource(LookupAddressUseCase lookupAddress) {
    this.lookupAddress = lookupAddress;
  }

  @GET
  @Path("/{cep}")
  public AddressResponse get(@PathParam("cep") String cep) {
    return AddressResponse.from(lookupAddress.lookup(cep));
  }
}
