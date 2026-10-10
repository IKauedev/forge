import { Link, Outlet } from 'react-router';

export default function Shell() {
  return (
    <>
      <header style={{ padding: '1rem 1.5rem', borderBottom: '1px solid #ddd' }}>
        <Link to="/" style={{ fontWeight: 600, textDecoration: 'none', color: 'inherit' }}>
          Starter App
        </Link>
      </header>
      <main style={{ padding: '1.5rem' }}>
        <Outlet />
      </main>
    </>
  );
}
