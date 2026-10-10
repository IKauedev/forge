package com.example.starterapi.adapter.in.web;

import com.example.starterapi.shared.security.Roles;
import jakarta.annotation.security.RolesAllowed;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;
import java.util.Map;

/** Exemplo de endpoint restrito a administradores. */
@Path("/api/admin")
@Produces(MediaType.APPLICATION_JSON)
@RolesAllowed(Roles.ADMIN)
public class AdminResource {

  @GET
  @Path("/ping")
  public Map<String, String> ping() {
    return Map.of("status", "ok");
  }
}
