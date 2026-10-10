import { useMemo, useSyncExternalStore } from 'react';
import { getConfig } from '../config/config';
import { decodeJwt, rolesFrom } from './jwt';
import { tokenStore } from './token-store';

export interface AuthState {
  /** false desliga os guards (config.json: auth.enabled). */
  enabled: boolean;
  authenticated: boolean;
  roles: string[];
  hasRole: (role: string) => boolean;
  hasAnyRole: (roles: readonly string[]) => boolean;
}

export function useAuth(): AuthState {
  const token = useSyncExternalStore(tokenStore.subscribe, tokenStore.get);
  const { auth } = getConfig();

  return useMemo(() => {
    const roles = token ? rolesFrom(decodeJwt(token), auth.rolesClaim) : [];
    const normalize = (role: string) => role.replace(/^ROLE_/i, '').toUpperCase();
    return {
      enabled: auth.enabled,
      authenticated: token !== null,
      roles,
      hasRole: (role) => roles.includes(normalize(role)),
      hasAnyRole: (wanted) => wanted.some((role) => roles.includes(normalize(role))),
    };
  }, [token, auth.enabled, auth.rolesClaim]);
}
