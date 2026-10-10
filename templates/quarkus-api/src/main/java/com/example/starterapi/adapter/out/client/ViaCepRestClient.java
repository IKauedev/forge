package com.example.starterapi.adapter.out.client;

import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.PathParam;
import org.eclipse.microprofile.rest.client.annotation.RegisterProvider;
import org.eclipse.microprofile.rest.client.inject.RegisterRestClient;

/** URL e timeouts em application.properties (quarkus.rest-client.viacep.*). */
@RegisterRestClient(configKey = "viacep")
@RegisterProvider(CorrelationIdClientFilter.class)
@Path("/{cep}/json")
public interface ViaCepRestClient {

  @GET
  ViaCepResponse find(@PathParam("cep") String cep);
}
