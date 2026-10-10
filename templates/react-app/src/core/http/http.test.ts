import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { tokenStore } from '../auth/token-store';
import { setConfig } from '../config/config';
import { useConfig } from '../../test/helpers';
import { apiFetch, ApiError, CORRELATION_HEADER } from './http';

const assign = vi.fn();

beforeEach(() => {
  useConfig();
  vi.stubGlobal('location', { assign });
});

afterEach(() => {
  vi.unstubAllGlobals();
  assign.mockReset();
  tokenStore.clear();
  setConfig(null);
});

function stubFetch(response: Response) {
  const fetchMock = vi.fn().mockResolvedValue(response);
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function sentHeaders(fetchMock: ReturnType<typeof vi.fn>): Headers {
  return (fetchMock.mock.calls[0][1] as RequestInit).headers as Headers;
}

describe('apiFetch', () => {
  it('chama a API configurada com o id de correlação', async () => {
    const fetchMock = stubFetch(Response.json({ ok: true }));

    await expect(apiFetch('/ping')).resolves.toEqual({ ok: true });

    expect(fetchMock.mock.calls[0][0]).toBe('https://api.example.com/ping');
    expect(sentHeaders(fetchMock).get(CORRELATION_HEADER)).toMatch(/^[0-9a-f-]{36}$/);
    expect(sentHeaders(fetchMock).has('Authorization')).toBe(false);
  });

  it('mantém o id de correlação informado e envia o Bearer quando há token', async () => {
    tokenStore.set('t0k3n');
    const fetchMock = stubFetch(Response.json({}));

    await apiFetch('/x', { headers: { [CORRELATION_HEADER]: 'abc' } });

    expect(sentHeaders(fetchMock).get(CORRELATION_HEADER)).toBe('abc');
    expect(sentHeaders(fetchMock).get('Authorization')).toBe('Bearer t0k3n');
  });

  it('lança ApiError com status e corpo', async () => {
    stubFetch(Response.json({ message: 'não' }, { status: 422 }));

    const error = await apiFetch('/x').catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 422, body: { message: 'não' } });
  });

  it('em 401 limpa o token e leva ao login', async () => {
    tokenStore.set('t0k3n');
    stubFetch(new Response('', { status: 401 }));

    await expect(apiFetch('/x')).rejects.toBeInstanceOf(ApiError);

    expect(tokenStore.get()).toBeNull();
    expect(assign).toHaveBeenCalledWith('/login');
  });

  it('em 401 com a autenticação desligada não redireciona', async () => {
    useConfig({ enabled: false });
    stubFetch(new Response('', { status: 401 }));

    await expect(apiFetch('/x')).rejects.toBeInstanceOf(ApiError);

    expect(assign).not.toHaveBeenCalled();
  });
});
