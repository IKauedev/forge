import { useEffect, useState, type ReactNode } from 'react';
import { login } from './login';
import { useAuth } from './use-auth';

interface GuardProps {
  children: ReactNode;
}

function Forbidden() {
  return <p role="alert">Você não tem permissão para acessar esta página.</p>;
}

/** Espelha @IsAuthenticated: exige usuário autenticado; senão leva ao login do provedor. */
export function RequireAuth({ children }: GuardProps) {
  const { enabled, authenticated } = useAuth();
  const blocked = enabled && !authenticated;
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (blocked) {
      login().catch((e: unknown) => setError(e instanceof Error ? e.message : 'Falha ao entrar.'));
    }
  }, [blocked]);

  if (error) return <p role="alert">{error}</p>;
  return blocked ? null : children;
}

/** Espelha @HasAnyRole: exige pelo menos uma das roles. */
export function RequireAnyRole({ roles, children }: GuardProps & { roles: readonly string[] }) {
  const auth = useAuth();
  return (
    <RequireAuth>{!auth.enabled || auth.hasAnyRole(roles) ? children : <Forbidden />}</RequireAuth>
  );
}

/** Espelha @HasRole. */
export function RequireRole({ role, children }: GuardProps & { role: string }) {
  return <RequireAnyRole roles={[role]}>{children}</RequireAnyRole>;
}

/** Espelha @IsAdmin. */
export function RequireAdmin({ children }: GuardProps) {
  return <RequireRole role="ADMIN">{children}</RequireRole>;
}
