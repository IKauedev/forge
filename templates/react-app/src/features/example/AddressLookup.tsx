import { useState, type FormEvent } from 'react';
import { lookupAddress, type Address } from './address-api';

export default function AddressLookup() {
  const [cep, setCep] = useState('');
  const [address, setAddress] = useState<Address | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function search(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError(null);
    setAddress(null);
    try {
      setAddress(await lookupAddress(cep));
    } catch {
      setError('Não foi possível consultar o CEP.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <h1>Consulta de CEP</h1>
      <form onSubmit={search}>
        <label htmlFor="cep">CEP</label>
        <input id="cep" name="cep" value={cep} onChange={(e) => setCep(e.target.value)} />
        <button type="submit" disabled={loading}>
          Buscar
        </button>
      </form>

      {loading && <p>Buscando…</p>}
      {error && <p role="alert">{error}</p>}
      {address && (
        <dl>
          <dt>Logradouro</dt>
          <dd>{address.street}</dd>
          <dt>Bairro</dt>
          <dd>{address.district}</dd>
          <dt>Cidade</dt>
          <dd>
            {address.city}/{address.state}
          </dd>
        </dl>
      )}
    </>
  );
}
