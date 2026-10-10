import { Component, inject } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';
import { AuthService } from '../core/auth/auth.service';
import { ConfigService } from '../core/config/config.service';

@Component({
  selector: 'app-shell',
  imports: [RouterLink, RouterOutlet],
  template: `
    <header>
      <a routerLink="/" class="brand">Starter App</a>
      @if (authEnabled && auth.isAuthenticated()) {
        <button type="button" (click)="auth.logout()">Sair</button>
      }
    </header>
    <main>
      <router-outlet />
    </main>
  `,
  styles: `
    header {
      display: flex;
      justify-content: space-between;
      padding: 1rem 1.5rem;
      border-bottom: 1px solid #ddd;
    }
    .brand {
      font-weight: 600;
      text-decoration: none;
      color: inherit;
    }
    main {
      padding: 1.5rem;
    }
  `,
})
export class Shell {
  protected readonly auth = inject(AuthService);
  protected readonly authEnabled = inject(ConfigService).config().auth.enabled;
}
