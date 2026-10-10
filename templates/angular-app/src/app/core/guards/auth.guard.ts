import { inject } from '@angular/core';
import { CanActivateFn } from '@angular/router';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';

/** Exige usuário autenticado; senão leva ao provedor de identidade e volta para a mesma página. */
export const authGuard: CanActivateFn = (_route, state) => {
  const { enabled } = inject(ConfigService).config().auth;
  const auth = inject(AuthService);
  if (!enabled || auth.isAuthenticated()) {
    return true;
  }
  auth.login(state.url).catch((e: unknown) => console.error('Falha ao iniciar o login', e));
  return false;
};
