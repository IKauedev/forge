import { useEffect, type ReactNode } from 'react';
import { getConfig } from '../config/config';
import { useAuth } from './use-auth';

interface GuardProps {
  children: ReactNode;
}

function Forbidden() {
  return <p role="alert">Você não tem permissão para acessar esta página.</p>;
}

/** Espelha @IsAuthenticated: exige usuário autenticado; senão leva para o login. */
export function RequireAuth({ children }: GuardProps) {
  const { enabled, authenticated } = useAuth();
  const blocked = enabled && !authenticated;

  useEffect(() => {
    if (blocked) window.location.assign(getConfig().auth.loginUrl);
  }, [blocked]);

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
