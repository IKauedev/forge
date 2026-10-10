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
        const { enabled, loginUrl } = config.config().auth;
        if (enabled) {
          window.location.assign(loginUrl);
        }
      }
      return throwError(() => error);
    }),
  );
};
