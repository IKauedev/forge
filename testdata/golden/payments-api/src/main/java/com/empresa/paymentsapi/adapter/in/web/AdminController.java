package com.empresa.paymentsapi.adapter.in.web;

import com.empresa.paymentsapi.shared.security.IsAdmin;
import java.util.Map;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** Exemplo de endpoint restrito a administradores com a anotação personalizada. */
@RestController
@RequestMapping("/api/admin")
@IsAdmin
public class AdminController {

  @GetMapping("/ping")
  public Map<String, String> ping() {
    return Map.of("status", "ok");
  }
}
