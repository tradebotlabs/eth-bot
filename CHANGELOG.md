# Changelog

All notable changes to ETH Trading Bot will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Email verification system
- Two-factor authentication (2FA)
- API key encryption at rest
- Password reset functionality
- Account lockout after failed attempts
- Advanced backtesting with walk-forward optimization
- Machine learning strategy integration

## [1.0.0] - TBD (MVP Release)

### Added

#### Authentication & User Management
- Multi-user authentication system with JWT tokens
- Secure password hashing using bcrypt (cost factor 12)
- Session management with refresh tokens (7-day expiry)
- Role-based access control (admin, trader, viewer)
- Demo and live trading account types
- User registration with account creation
- Protected routes with authentication middleware
- Automatic token refresh on expiration
- Logout functionality with session invalidation
- Audit logging for security events

#### Database
- PostgreSQL integration for user and authentication data
- Database migration system with up/down migrations
- Seed data for development and testing
- Docker Compose setup for PostgreSQL
- Automated schema initialization
- Session storage with IP and user agent tracking

#### Frontend
- Login page with form validation and error handling
- Registration page with demo/live account selection
- Demo account capital slider ($1,000 - $100,000)
- Live account Binance API credential inputs
- Protected route wrapper component
- Axios interceptors for automatic auth headers
- Token persistence with localStorage
- User profile display in sidebar
- Logout button with confirmation
- Responsive mobile-first design
- Dark mode support

#### Trading Engine
- 5 advanced trading strategies:
  - Trend Following (EMA crossovers, ADX confirmation)
  - Mean Reversion (RSI-based with Bollinger Bands)
  - Breakout Trading (Volume-confirmed breakouts)
  - Volatility Trading (ATR-based adaptive strategies)
  - Statistical Arbitrage (Pair trading and market neutral)
- Paper trading mode with realistic simulation
- Live trading with Binance integration
- Multi-timeframe analysis (1m, 5m, 15m, 1h, 4h, 1d)
- Real-time WebSocket data streaming
- Order management (Market, Limit, Stop-Loss)
- Position tracking and P&L calculation

#### Risk Management
- Position sizing with multiple methods:
  - Fixed fractional allocation
  - Risk-adjusted sizing based on volatility
  - Maximum position size limits
- Risk controls:
  - Per-trade stop-loss limits
  - Maximum drawdown protection (20% default)
  - Daily loss limits (5% default)
  - Weekly loss limits (10% default)
  - Maximum open positions limit
  - Leverage controls
- Circuit breaker system:
  - Automatic trading halt after consecutive losses
  - Configurable halt duration
  - Manual resume capability

#### Technical Indicators
- Moving Averages (SMA, EMA, WMA)
- Momentum Indicators (RSI, Stochastic, MACD)
- Volatility Metrics (Bollinger Bands, ATR, ADX)
- Volume Analysis (OBV, Volume Profile)
- Configurable indicator parameters

#### Dashboard & UI
- Real-time TradingView-style charts
- Live price ticker with WebSocket updates
- Account statistics footer:
  - Current balance and equity
  - Unrealized P&L
  - Daily P&L
  - Win rate percentage
  - Current drawdown
- Trading mode indicator (Paper/Live)
- Start/Stop trading controls
- Connection status indicator
- Collapsible sidebar navigation
- Mobile-responsive design
- Toast notifications for events

#### Developer Experience
- Comprehensive README with quick start guide
- CONTRIBUTING.md with development guidelines
- SECURITY.md with vulnerability reporting process
- Database migration documentation
- Docker Compose for local development
- Seed data for testing
- Example configuration file
- Code of conduct

#### Infrastructure
- Go 1.22+ backend with Echo framework
- React 19 frontend with TypeScript
- PostgreSQL 16+ for user data
- SQLite for trading data
- Docker containerization support
- Health check endpoints
- Structured logging with zerolog
- CORS configuration

### Changed
- Migrated authentication from in-memory to PostgreSQL
- Updated configuration system to support auth settings
- Enhanced API server with protected routes
- Improved error handling and validation
- Updated frontend routing for authentication flow

### Security
- Implemented JWT-based authentication
- Added bcrypt password hashing
- Session tracking with IP and user agent
- Protected API endpoints with middleware
- Secure token storage in localStorage
- Automatic token refresh mechanism
- SQL injection prevention with parameterized queries
- Input validation on all endpoints
- Rate limiting recommendations
- Security audit logging

### Documentation
- Added PostgreSQL setup instructions
- Documented demo vs live account types
- Added seed data credentials
- Updated architecture diagrams
- Enhanced API documentation
- Added security best practices
- Migration guide for database changes

### Fixed
- Repository naming conflicts (AccountRepository → TradingAccountRepository)
- Build errors in authentication system integration
- CORS issues with frontend authentication
- Token expiration handling
- Protected route navigation

## [0.9.0] - Initial Development Release

### Added
- Basic trading engine with strategy framework
- Paper trading executor
- Binance WebSocket client
- SQLite database for trading data
- React dashboard with basic charts
- Configuration system
- Orchestrator event loop

### Notes
- Pre-authentication version
- Single-user mode only
- Basic risk management
- Limited backtesting capabilities

---

## Release Notes Format

### Version Numbering
- **Major (X.0.0)**: Breaking changes, major features
- **Minor (0.X.0)**: New features, backwards compatible
- **Patch (0.0.X)**: Bug fixes, minor improvements

### Change Categories
- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security improvements

---

*For full commit history, see: https://github.com/yourusername/eth-trading/commits/master*

[Unreleased]: https://github.com/yourusername/eth-trading/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yourusername/eth-trading/releases/tag/v1.0.0
[0.9.0]: https://github.com/yourusername/eth-trading/releases/tag/v0.9.0
