import { describe, expect, it } from 'vitest';
import { fakeJwt } from '../../test/helpers';
import { decodeJwt, rolesFrom } from './jwt';

describe('jwt', () => {
  it('lê as claims do payload, inclusive com acentos', () => {
    const claims = decodeJwt(fakeJwt({ sub: '1', name: 'José' }));
    expect(claims).toMatchObject({ sub: '1' });
    expect(claims?.name).toBe('José');
  });

  it('devolve null para token inválido', () => {
    expect(decodeJwt('lixo')).toBeNull();
    expect(decodeJwt('a.%%%.c')).toBeNull();
  });

  it('normaliza as roles como o backend: sem ROLE_ e em maiúsculas', () => {
    expect(rolesFrom({ roles: ['admin', 'ROLE_user'] }, 'roles')).toEqual(['ADMIN', 'USER']);
  });

  it('aceita o claim do Cognito e roles em string', () => {
    expect(rolesFrom({ 'cognito:groups': ['Admin'] }, 'cognito:groups')).toEqual(['ADMIN']);
    expect(rolesFrom({ roles: 'a b,c' }, 'roles')).toEqual(['A', 'B', 'C']);
  });

  it('sem o claim não há roles', () => {
    expect(rolesFrom({}, 'roles')).toEqual([]);
    expect(rolesFrom(null, 'roles')).toEqual([]);
  });
});
