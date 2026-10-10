import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';

/**
 * Anexa o Bearer token somente nas chamadas para a API configurada. Usa `current()`: a requisição que
 * carrega o config.json passa por aqui antes de a configuração existir.
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const config = inject(ConfigService).current();
  const token = inject(AuthService).token();
  if (!config || !token || !req.url.startsWith(config.apiBaseUrl)) {
    return next(req);
  }
  return next(req.clone({ setHeaders: { Authorization: `Bearer ${token}` } }));
};
