import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { AddressLookup } from './address-lookup';
import { Address, AddressService } from './address.service';

const address: Address = {
  cep: '01310100',
  street: 'Avenida Paulista',
  district: 'Bela Vista',
  city: 'São Paulo',
  state: 'SP',
};

function setup(lookup: AddressService['lookup']) {
  TestBed.configureTestingModule({
    imports: [AddressLookup],
    providers: [{ provide: AddressService, useValue: { lookup } }],
  });
  const fixture = TestBed.createComponent(AddressLookup);
  fixture.detectChanges();
  return fixture;
}

describe('AddressLookup', () => {
  it('mostra o endereço encontrado', () => {
    const fixture = setup(() => of(address));
    fixture.componentInstance.cep.set('01310-100');
    fixture.componentInstance.search();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Avenida Paulista');
    expect(fixture.nativeElement.textContent).toContain('São Paulo/SP');
  });

  it('mostra uma mensagem quando a consulta falha', () => {
    const fixture = setup(() => throwError(() => new Error('boom')));
    fixture.componentInstance.search();
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('[role="alert"]')?.textContent).toContain(
      'Não foi possível consultar o CEP.',
    );
  });
});
