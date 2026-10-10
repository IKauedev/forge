package com.example.starterapi.adapter.in.web;

import com.example.starterapi.shared.security.CurrentUser;
import io.quarkus.security.Authenticated;
import io.quarkus.security.identity.SecurityIdentity;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;

@Path("/api/me")
@Produces(MediaType.APPLICATION_JSON)
@Authenticated
public class MeResource {

  private final SecurityIdentity identity;

  public MeResource(SecurityIdentity identity) {
    this.identity = identity;
  }

  @GET
  public CurrentUser me() {
    return CurrentUser.from(identity);
  }
}
