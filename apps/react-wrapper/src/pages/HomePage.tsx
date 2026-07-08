import { Link } from 'react-router-dom';

export function HomePage() {
  return (
    <div style={{ padding: '2rem', fontFamily: 'system-ui, sans-serif', height: '100%' }}>
      <h1>react-wrapper</h1>
      <p>
        <Link to="/about">Open embedded microfrontend</Link>
      </p>
    </div>
  );
}
