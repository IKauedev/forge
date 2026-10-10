import { HttpInterceptorFn } from '@angular/common/http';

export const CORRELATION_HEADER = 'X-Correlation-Id';

export const correlationIdInterceptor: HttpInterceptorFn = (req, next) => {
  if (req.headers.has(CORRELATION_HEADER)) {
    return next(req);
  }
  return next(req.clone({ setHeaders: { [CORRELATION_HEADER]: crypto.randomUUID() } }));
};
