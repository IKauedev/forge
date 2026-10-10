package com.empresa.ordersservice.adapter.in.web;

import com.empresa.ordersservice.shared.security.CurrentUser;
import com.empresa.ordersservice.shared.security.IsAuthenticated;
import com.empresa.ordersservice.shared.security.SecurityUtils;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/me")
public class MeController {

  @GetMapping
  @IsAuthenticated
  public CurrentUser me() {
    return SecurityUtils.currentUser().orElseThrow();
  }
}
