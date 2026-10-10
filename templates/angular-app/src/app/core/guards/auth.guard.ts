import { inject } from '@angular/core';
import { CanActivateFn } from '@angular/router';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';

export const authGuard: CanActivateFn = () => {
  const { enabled, loginUrl } = inject(ConfigService).config().auth;
  if (!enabled || inject(AuthService).isAuthenticated()) {
    return true;
  }
  window.location.assign(loginUrl);
  return false;
};
