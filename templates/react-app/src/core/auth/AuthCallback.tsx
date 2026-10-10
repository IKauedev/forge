import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { finishLogin, login } from './login';

/** Rota /auth/callback: termina o login OIDC e volta para a página que o usuário queria. */
export default function AuthCallback() {
  const navigate = useNavigate();
  const started = useRef(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // O StrictMode roda o efeito duas vezes em dev, e o code só pode ser trocado uma vez.
    if (started.current) return;
    started.current = true;
    finishLogin()
      .then((returnUrl) => navigate(returnUrl, { replace: true }))
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'Falha ao entrar.'));
  }, [navigate]);

  if (error) {
    return (
      <>
        <h1>Não foi possível entrar</h1>
        <p role="alert">{error}</p>
        <button type="button" onClick={() => void login('/')}>
          Tentar de novo
        </button>
      </>
    );
  }
  return <p>Entrando…</p>;
}
