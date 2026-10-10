import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { setConfig } from '../../core/config/config';
import { useConfig } from '../../test/helpers';
import AddressLookup from './AddressLookup';

beforeEach(() => {
  useConfig({ enabled: false });
});

afterEach(() => {
  vi.unstubAllGlobals();
  setConfig(null);
});

describe('AddressLookup', () => {
  it('consulta só os dígitos do CEP e mostra o endereço', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      Response.json({
        cep: '01310100',
        street: 'Avenida Paulista',
        district: 'Bela Vista',
        city: 'São Paulo',
        state: 'SP',
      }),
    );
    vi.stubGlobal('fetch', fetchMock);

    render(<AddressLookup />);
    await userEvent.type(screen.getByLabelText('CEP'), '01310-100');
    await userEvent.click(screen.getByRole('button', { name: 'Buscar' }));

    expect(await screen.findByText('Avenida Paulista')).toBeInTheDocument();
    expect(screen.getByText('São Paulo/SP')).toBeInTheDocument();
    expect(fetchMock.mock.calls[0][0]).toBe('https://api.example.com/addresses/01310100');
  });

  it('mostra uma mensagem quando a consulta falha', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 500 })));

    render(<AddressLookup />);
    await userEvent.click(screen.getByRole('button', { name: 'Buscar' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível consultar o CEP.');
  });
});
