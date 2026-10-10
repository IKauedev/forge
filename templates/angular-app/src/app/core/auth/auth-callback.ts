import { Component, OnInit, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from './auth.service';

/** Rota /auth/callback: termina o login OIDC e volta para a página que o usuário queria. */
@Component({
  selector: 'app-auth-callback',
  template: `
    @if (error(); as message) {
      <h1>Não foi possível entrar</h1>
      <p role="alert">{{ message }}</p>
      <button type="button" (click)="retry()">Tentar de novo</button>
    } @else {
      <p>Entrando…</p>
    }
  `,
})
export class AuthCallback implements OnInit {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly error = signal<string | null>(null);

  ngOnInit(): void {
    this.auth
      .completeLogin()
      .then((returnUrl) => this.router.navigateByUrl(returnUrl, { replaceUrl: true }))
      .catch((e: unknown) => this.error.set(e instanceof Error ? e.message : 'Falha ao entrar.'));
  }

  retry(): void {
    void this.auth.login('/');
  }
}
