import { Component } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-shell',
  imports: [RouterLink, RouterOutlet],
  template: `
    <header>
      <a routerLink="/" class="brand">Starter App</a>
    </header>
    <main>
      <router-outlet />
    </main>
  `,
  styles: `
    header {
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
export class Shell {}
