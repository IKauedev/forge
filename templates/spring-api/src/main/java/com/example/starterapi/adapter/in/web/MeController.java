package com.example.starterapi.adapter.in.web;

import com.example.starterapi.shared.security.CurrentUser;
import com.example.starterapi.shared.security.IsAuthenticated;
import com.example.starterapi.shared.security.SecurityUtils;
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
