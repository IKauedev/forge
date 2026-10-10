package com.example.starterapi.config;

import jakarta.ws.rs.core.Application;
import org.eclipse.microprofile.openapi.annotations.OpenAPIDefinition;
import org.eclipse.microprofile.openapi.annotations.enums.SecuritySchemeType;
import org.eclipse.microprofile.openapi.annotations.info.Info;
import org.eclipse.microprofile.openapi.annotations.security.SecurityScheme;

// forge:if auth!=none
@SecurityScheme(
    securitySchemeName = "bearer",
    type = SecuritySchemeType.HTTP,
    scheme = "bearer",
    bearerFormat = "JWT")
// forge:end
@OpenAPIDefinition(info = @Info(title = "Starter Api", version = "v1"))
public class OpenApiConfig extends Application {}
