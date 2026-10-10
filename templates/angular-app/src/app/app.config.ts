import {
  ApplicationConfig,
  inject,
  provideAppInitializer,
  provideBrowserGlobalErrorListeners,
} from '@angular/core';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideRouter, withComponentInputBinding } from '@angular/router';
import { routes } from './app.routes';
import { ConfigService } from './core/config/config.service';
import { authInterceptor } from './core/interceptors/auth.interceptor';
import { correlationIdInterceptor } from './core/interceptors/correlation-id.interceptor';
import { errorInterceptor } from './core/interceptors/error.interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes, withComponentInputBinding()),
    // A ordem importa: o id de correlação e o token entram antes de o erro ser tratado.
    provideHttpClient(
      withInterceptors([correlationIdInterceptor, authInterceptor, errorInterceptor]),
    ),
    // Carrega /config.json antes de renderizar: a mesma build roda em qualquer ambiente.
    provideAppInitializer(() => inject(ConfigService).load()),
  ],
};
