import { afterEach, describe, expect, it, vi } from 'vitest';
import { getConfig, loadConfig, setConfig } from './config';

afterEach(() => {
  vi.unstubAllGlobals();
  setConfig(null);
});

describe('config', () => {
  it('falha ao ler a configuração antes de carregar', () => {
    expect(() => getConfig()).toThrow();
  });

  it('carrega o config.json sem cache e expõe o valor', async () => {
    const body = { apiBaseUrl: 'https://api.example.com', auth: { enabled: true } };
    const fetchMock = vi.fn().mockResolvedValue(Response.json(body));
    vi.stubGlobal('fetch', fetchMock);

    await loadConfig();

    expect(fetchMock).toHaveBeenCalledWith('/config.json', { cache: 'no-store' });
    expect(getConfig().apiBaseUrl).toBe('https://api.example.com');
  });

  it('rejeita quando o config.json não existe', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 404 })));
    await expect(loadConfig()).rejects.toThrow('404');
  });
});
