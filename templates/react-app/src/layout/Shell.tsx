import { Link, Outlet } from 'react-router';
import { logout } from '../core/auth/login';
import { useAuth } from '../core/auth/use-auth';

export default function Shell() {
  const { enabled, authenticated } = useAuth();
  return (
    <>
      <header
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          padding: '1rem 1.5rem',
          borderBottom: '1px solid #ddd',
        }}
      >
        <Link to="/" style={{ fontWeight: 600, textDecoration: 'none', color: 'inherit' }}>
          Starter App
        </Link>
        {enabled && authenticated && (
          <button type="button" onClick={() => void logout()}>
            Sair
          </button>
        )}
      </header>
      <main style={{ padding: '1.5rem' }}>
        <Outlet />
      </main>
    </>
  );
}
