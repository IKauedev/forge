import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const config = inject(ConfigService);
  return next(req).pipe(
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse && error.status === 401) {
        auth.clear();
        // sessão vencida ou revogada: volta pelo provedor e depois para a mesma página
        if (config.current()?.auth.enabled) {
          auth.login().catch(() => undefined);
        }
      }
      return throwError(() => error);
    }),
  );
};
