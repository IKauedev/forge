import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-not-found',
  imports: [RouterLink],
  template: `
    <h1>Página não encontrada</h1>
    <a routerLink="/">Voltar ao início</a>
  `,
})
export class NotFound {}
