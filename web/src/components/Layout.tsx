import { Link, useLocation, Outlet, useNavigate } from 'react-router-dom';
import { useTradingStore } from '../stores/tradingStore';
import { useAuthStore } from '../stores/authStore';
import { useWebSocket } from '../hooks/useWebSocket';
import { useState, useEffect } from 'react';
import * as api from '../services/api';

const navItems = [
  { path: '/', label: 'Dashboard', icon: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
  )},
  { path: '/strategies', label: 'Strategies', icon: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M12 2L2 7l10 5 10-5-10-5z" />
      <path d="M2 17l10 5 10-5" />
      <path d="M2 12l10 5 10-5" />
    </svg>
  )},
  { path: '/risk', label: 'Risk', icon: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    </svg>
  )},
  { path: '/backtest', label: 'Backtest', icon: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
    </svg>
  )},
  { path: '/analytics', label: 'Analytics', icon: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M21.21 15.89A10 10 0 1 1 8 2.83" />
      <path d="M22 12A10 10 0 0 0 12 2v10z" />
    </svg>
  )},
  { path: '/settings', label: 'Settings', icon: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="3" />
      <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
    </svg>
  )},
];

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  useWebSocket();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(true);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const [isToggling, setIsToggling] = useState(false);

  const { wsConnected, isRunning, mode, currentPrice, accountStats, setStatus } = useTradingStore();
  const { user, logout } = useAuthStore();

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  // Check for mobile viewport
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth <= 768);
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // Close mobile menu on route change
  useEffect(() => {
    setMobileMenuOpen(false);
  }, [location.pathname]);

  // Prevent body scroll when mobile menu is open
  useEffect(() => {
    if (mobileMenuOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [mobileMenuOpen]);

  const handleToggleTrading = async () => {
    if (isToggling) return;
    setIsToggling(true);
    try {
      if (isRunning) {
        await api.stopTrading();
        setStatus({ running: false });
      } else {
        await api.startTrading();
        setStatus({ running: true });
      }
    } catch (error) {
      console.error('Failed to toggle trading:', error);
    } finally {
      setIsToggling(false);
    }
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
    }).format(price);
  };

  const formatPriceCompact = (price: number) => {
    if (price >= 1000) {
      return `$${(price / 1000).toFixed(1)}k`;
    }
    return formatPrice(price);
  };

  // Sidebar content (shared between desktop and mobile)
  const SidebarContent = ({ expanded }: { expanded: boolean }) => (
    <>
      {/* Logo */}
      <div style={{
        height: '44px',
        borderBottom: '1px solid #2a2e39',
        display: 'flex',
        alignItems: 'center',
        justifyContent: expanded ? 'flex-start' : 'center',
        padding: '0 12px',
        gap: '8px',
      }}>
        <div
          style={{
            width: '28px',
            height: '28px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: '#2962ff',
            borderRadius: '6px',
            color: 'white',
            fontSize: '11px',
            fontWeight: 700,
            flexShrink: 0,
          }}
        >
          ET
        </div>
        {expanded && (
          <span style={{
            fontWeight: 600,
            fontSize: '13px',
            whiteSpace: 'nowrap',
            color: '#d1d4dc',
          }}>
            ETH Trader
          </span>
        )}
        {expanded && isMobile && (
          <button
            onClick={() => setMobileMenuOpen(false)}
            style={{
              marginLeft: 'auto',
              background: 'transparent',
              border: 'none',
              color: '#787b86',
              cursor: 'pointer',
              padding: '4px',
            }}
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>

      {/* Navigation */}
      <nav style={{ flex: 1, padding: '6px 4px', overflowY: 'auto' }}>
        {navItems.map((item) => (
          <Link
            key={item.path}
            to={item.path}
            title={!expanded ? item.label : undefined}
            onClick={() => isMobile && setMobileMenuOpen(false)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              padding: expanded ? '10px 12px' : '8px',
              marginBottom: '2px',
              borderRadius: '6px',
              color: location.pathname === item.path ? '#d1d4dc' : '#787b86',
              background: location.pathname === item.path ? '#2a2e39' : 'transparent',
              textDecoration: 'none',
              fontSize: '13px',
              fontWeight: location.pathname === item.path ? 500 : 400,
              transition: 'all 0.1s ease',
              justifyContent: expanded ? 'flex-start' : 'center',
            }}
          >
            {item.icon}
            {expanded && <span>{item.label}</span>}
          </Link>
        ))}
      </nav>

      {/* User & Status */}
      <div style={{
        borderTop: '1px solid #2a2e39',
      }}>
        {/* User Info */}
        {expanded && user && (
          <div style={{
            padding: '12px',
            borderBottom: '1px solid #2a2e39',
          }}>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              marginBottom: '8px',
            }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '50%',
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'white',
                fontSize: '13px',
                fontWeight: 600,
                flexShrink: 0,
              }}>
                {user.full_name.charAt(0).toUpperCase()}
              </div>
              <div style={{ minWidth: 0, flex: 1 }}>
                <div style={{
                  color: '#d1d4dc',
                  fontSize: '12px',
                  fontWeight: 500,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}>
                  {user.full_name}
                </div>
                <div style={{
                  color: '#787b86',
                  fontSize: '10px',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}>
                  {user.email}
                </div>
              </div>
            </div>
            <button
              onClick={handleLogout}
              style={{
                width: '100%',
                padding: '6px 8px',
                fontSize: '11px',
                fontWeight: 500,
                color: '#787b86',
                background: 'transparent',
                border: '1px solid #2a2e39',
                borderRadius: '4px',
                cursor: 'pointer',
                transition: 'all 0.1s ease',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '6px',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = '#2a2e39';
                e.currentTarget.style.color = '#d1d4dc';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'transparent';
                e.currentTarget.style.color = '#787b86';
              }}
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
              Logout
            </button>
          </div>
        )}

        {/* Connection Status */}
        <div style={{
          padding: '12px',
          fontSize: '11px',
        }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: expanded ? 'flex-start' : 'center',
            gap: '6px',
          }}>
            <div style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              background: wsConnected ? '#26a69a' : '#ef5350',
              boxShadow: wsConnected ? '0 0 6px #26a69a' : '0 0 6px #ef5350',
            }} />
            {expanded && (
              <span style={{ color: '#787b86', fontSize: '11px' }}>
                {wsConnected ? 'Connected' : 'Offline'}
              </span>
            )}
          </div>
        </div>
      </div>
    </>
  );

  return (
    <div style={{
      display: 'flex',
      height: '100vh',
      overflow: 'hidden',
      background: '#131722'
    }}>
      {/* Mobile Overlay */}
      {mobileMenuOpen && (
        <div
          className="mobile-overlay active"
          onClick={() => setMobileMenuOpen(false)}
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0, 0, 0, 0.6)',
            zIndex: 40,
          }}
        />
      )}

      {/* Desktop Sidebar */}
      {!isMobile && (
        <aside
          className="desktop-sidebar"
          style={{
            width: sidebarCollapsed ? '48px' : '180px',
            background: '#1e222d',
            borderRight: '1px solid #2a2e39',
            display: 'flex',
            flexDirection: 'column',
            transition: 'width 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
            flexShrink: 0,
          }}
          onMouseEnter={() => setSidebarCollapsed(false)}
          onMouseLeave={() => setSidebarCollapsed(true)}
        >
          <SidebarContent expanded={!sidebarCollapsed} />
        </aside>
      )}

      {/* Mobile Sidebar */}
      {isMobile && (
        <aside
          className={`sidebar-mobile ${mobileMenuOpen ? 'open' : ''}`}
          style={{
            position: 'fixed',
            left: 0,
            top: 0,
            bottom: 0,
            width: '260px',
            background: '#1e222d',
            borderRight: '1px solid #2a2e39',
            display: 'flex',
            flexDirection: 'column',
            zIndex: 50,
            transform: mobileMenuOpen ? 'translateX(0)' : 'translateX(-100%)',
            transition: 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          }}
        >
          <SidebarContent expanded={true} />
        </aside>
      )}

      {/* Main */}
      <main style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        height: '100vh',
        minWidth: 0,
      }}>
        {/* Top bar */}
        <header
          className="header-mobile"
          style={{
            height: isMobile ? '48px' : '44px',
            background: '#1e222d',
            borderBottom: '1px solid #2a2e39',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 12px',
            flexShrink: 0,
            gap: '8px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px', minWidth: 0, flex: 1 }}>
            {/* Mobile Menu Button */}
            {isMobile && (
              <button
                className="mobile-menu-btn"
                onClick={() => setMobileMenuOpen(true)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: '36px',
                  height: '36px',
                  background: 'transparent',
                  border: 'none',
                  color: '#787b86',
                  cursor: 'pointer',
                  borderRadius: '6px',
                  flexShrink: 0,
                }}
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M3 12h18M3 6h18M3 18h18" />
                </svg>
              </button>
            )}

            {currentPrice > 0 && (
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                minWidth: 0,
              }}>
                <span style={{
                  fontSize: isMobile ? '9px' : '10px',
                  color: '#787b86',
                  textTransform: 'uppercase',
                  letterSpacing: '0.03em',
                  flexShrink: 0,
                }}>
                  ETH/USDT
                </span>
                <span style={{
                  fontSize: isMobile ? '13px' : '14px',
                  fontWeight: 600,
                  color: '#d1d4dc',
                  flexShrink: 0,
                }}>
                  {formatPrice(currentPrice)}
                </span>
              </div>
            )}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', flexShrink: 0 }}>
            <span style={{
              padding: '2px 6px',
              fontSize: isMobile ? '9px' : '10px',
              fontWeight: 500,
              borderRadius: '3px',
              background: mode === 'live' ? 'rgba(255, 152, 0, 0.2)' : 'rgba(41, 98, 255, 0.2)',
              color: mode === 'live' ? '#ff9800' : '#2962ff',
            }}>
              {mode.toUpperCase()}
            </span>
            <button
              onClick={handleToggleTrading}
              disabled={isToggling}
              style={{
                padding: isMobile ? '6px 10px' : '4px 10px',
                fontSize: isMobile ? '11px' : '11px',
                fontWeight: 500,
                border: 'none',
                borderRadius: '3px',
                cursor: isToggling ? 'not-allowed' : 'pointer',
                opacity: isToggling ? 0.7 : 1,
                background: isRunning ? 'rgba(239, 83, 80, 0.2)' : 'rgba(38, 166, 154, 0.2)',
                color: isRunning ? '#ef5350' : '#26a69a',
              }}
            >
              {isToggling ? '...' : isRunning ? 'Stop' : 'Start'}
            </button>
          </div>
        </header>

        {/* Content */}
        <div style={{
          flex: 1,
          overflow: 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}>
          <Outlet />
        </div>

        {/* Footer Bar - Account Stats */}
        <footer
          className="footer-mobile"
          style={{
            height: isMobile ? '32px' : '28px',
            background: '#1e222d',
            borderTop: '1px solid #2a2e39',
            display: 'flex',
            alignItems: 'center',
            justifyContent: isMobile ? 'flex-start' : 'center',
            gap: isMobile ? '10px' : '16px',
            flexShrink: 0,
            fontSize: isMobile ? '10px' : '11px',
            padding: isMobile ? '0 12px' : '0 12px',
            overflowX: isMobile ? 'auto' : 'visible',
            whiteSpace: 'nowrap',
          }}
        >
          {accountStats ? (
            <>
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px', flexShrink: 0 }}>
                <span style={{ color: '#787b86' }}>Capital</span>
                <span style={{ color: '#d1d4dc', fontWeight: 500 }}>
                  {isMobile ? formatPriceCompact(accountStats.balance) : formatPrice(accountStats.balance)}
                </span>
              </div>
              <div style={{ width: '1px', height: '12px', background: '#2a2e39', flexShrink: 0 }} />
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px', flexShrink: 0 }}>
                <span style={{ color: '#787b86' }}>Equity</span>
                <span style={{ color: '#d1d4dc', fontWeight: 500 }}>
                  {isMobile ? formatPriceCompact(accountStats.equity) : formatPrice(accountStats.equity)}
                </span>
              </div>
              <div style={{ width: '1px', height: '12px', background: '#2a2e39', flexShrink: 0 }} />
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px', flexShrink: 0 }}>
                <span style={{ color: '#787b86' }}>{isMobile ? 'PnL' : 'Unrealized'}</span>
                <span style={{ fontWeight: 500, color: accountStats.unrealizedPnl >= 0 ? '#26a69a' : '#ef5350' }}>
                  {accountStats.unrealizedPnl >= 0 ? '+' : ''}{isMobile ? formatPriceCompact(accountStats.unrealizedPnl) : formatPrice(accountStats.unrealizedPnl)}
                </span>
              </div>
              {!isMobile && (
                <>
                  <div style={{ width: '1px', height: '12px', background: '#2a2e39' }} />
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span style={{ color: '#787b86' }}>Daily P&L</span>
                    <span style={{ fontWeight: 500, color: accountStats.dailyPnl >= 0 ? '#26a69a' : '#ef5350' }}>
                      {accountStats.dailyPnl >= 0 ? '+' : ''}{formatPrice(accountStats.dailyPnl)}
                    </span>
                  </div>
                </>
              )}
              <div style={{ width: '1px', height: '12px', background: '#2a2e39', flexShrink: 0 }} />
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px', flexShrink: 0 }}>
                <span style={{ color: '#787b86' }}>Win</span>
                <span style={{ color: '#d1d4dc', fontWeight: 500 }}>{(accountStats.winRate * 100).toFixed(0)}%</span>
              </div>
              <div style={{ width: '1px', height: '12px', background: '#2a2e39', flexShrink: 0 }} />
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px', flexShrink: 0 }}>
                <span style={{ color: '#787b86' }}>DD</span>
                <span style={{ fontWeight: 500, color: accountStats.currentDrawdown > 0.05 ? '#ef5350' : '#ff9800' }}>
                  {(accountStats.currentDrawdown * 100).toFixed(1)}%
                </span>
              </div>
            </>
          ) : (
            <span style={{ color: '#787b86' }}>Waiting for account data...</span>
          )}
        </footer>
      </main>
    </div>
  );
}
