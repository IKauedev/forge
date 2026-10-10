import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';

/** Anexa o Bearer token somente nas chamadas para a API configurada. */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const token = inject(AuthService).token();
  const { apiBaseUrl } = inject(ConfigService).config();
  if (!token || !req.url.startsWith(apiBaseUrl)) {
    return next(req);
  }
  return next(req.clone({ setHeaders: { Authorization: `Bearer ${token}` } }));
};
