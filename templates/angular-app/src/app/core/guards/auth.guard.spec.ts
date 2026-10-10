import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';
import { authGuard } from './auth.guard';

function run(enabled: boolean, token: string | null): boolean {
  sessionStorage.clear();
  TestBed.configureTestingModule({
    providers: [
      {
        provide: ConfigService,
        useValue: { config: () => ({ apiBaseUrl: '', auth: { enabled, loginUrl: '/login' } }) },
      },
    ],
  });
  if (token) {
    TestBed.inject(AuthService).setToken(token);
  }
  return TestBed.runInInjectionContext(
    () => authGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot) as boolean,
  );
}

describe('authGuard', () => {
  it('libera quando a autenticação está desligada', () => {
    expect(run(false, null)).toBe(true);
  });

  it('libera com token quando a autenticação está ligada', () => {
    expect(run(true, 'abc')).toBe(true);
  });
});
