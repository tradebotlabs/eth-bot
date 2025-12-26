import { useEffect } from 'react';
import { Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Strategies } from './pages/Strategies';
import { Risk } from './pages/Risk';
import { Backtest } from './pages/Backtest';
import { Analytics } from './pages/Analytics';
import { Settings } from './pages/Settings';

function App() {
  // Enable dark mode by default for trading platform
  useEffect(() => {
    document.documentElement.classList.add('dark');
  }, []);

  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="strategies" element={<Strategies />} />
        <Route path="risk" element={<Risk />} />
        <Route path="backtest" element={<Backtest />} />
        <Route path="analytics" element={<Analytics />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  );
}

export default App;
