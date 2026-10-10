export type JwtClaims = Record<string, unknown>;

/**
 * Lê as claims de um JWT SEM validar a assinatura. Serve só para decidir o que mostrar na tela:
 * a autorização de verdade é do backend (@IsAuthenticated, @HasRole etc.).
 */
export function decodeJwt(token: string): JwtClaims | null {
  const payload = token.split('.')[1];
  if (!payload) return null;
  try {
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/');
    const bytes = Uint8Array.from(atob(base64), (c) => c.charCodeAt(0));
    const parsed: unknown = JSON.parse(new TextDecoder().decode(bytes));
    return typeof parsed === 'object' && parsed !== null ? (parsed as JwtClaims) : null;
  } catch {
    return null;
  }
}

/** Roles em maiúsculas e sem prefixo ROLE_, como o backend as enxerga. */
export function rolesFrom(claims: JwtClaims | null, claim: string): string[] {
  const value = claims?.[claim];
  const list = Array.isArray(value)
    ? value
    : typeof value === 'string'
      ? value.split(/[\s,]+/)
      : [];
  return list
    .filter((r): r is string => typeof r === 'string' && r !== '')
    .map((r) => r.replace(/^ROLE_/i, '').toUpperCase());
}
