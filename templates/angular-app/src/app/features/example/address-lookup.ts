import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Address, AddressService } from './address.service';

@Component({
  selector: 'app-address-lookup',
  imports: [FormsModule],
  template: `
    <h1>Consulta de CEP</h1>
    <form (ngSubmit)="search()">
      <label for="cep">CEP</label>
      <input id="cep" name="cep" [ngModel]="cep()" (ngModelChange)="cep.set($event)" />
      <button type="submit" [disabled]="loading()">Buscar</button>
    </form>

    @if (loading()) {
      <p>Buscando…</p>
    }
    @if (error(); as message) {
      <p role="alert">{{ message }}</p>
    }
    @if (address(); as a) {
      <dl>
        <dt>Logradouro</dt>
        <dd>{{ a.street }}</dd>
        <dt>Bairro</dt>
        <dd>{{ a.district }}</dd>
        <dt>Cidade</dt>
        <dd>{{ a.city }}/{{ a.state }}</dd>
      </dl>
    }
  `,
})
export class AddressLookup {
  private readonly service = inject(AddressService);

  readonly cep = signal('');
  readonly address = signal<Address | null>(null);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);

  search(): void {
    this.loading.set(true);
    this.error.set(null);
    this.address.set(null);
    this.service.lookup(this.cep()).subscribe({
      next: (a) => {
        this.address.set(a);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Não foi possível consultar o CEP.');
        this.loading.set(false);
      },
    });
  }
}
