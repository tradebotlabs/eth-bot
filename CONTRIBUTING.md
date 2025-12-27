# Contributing to ETH Trading Bot

Thank you for your interest in contributing to the ETH Trading Bot project! We welcome contributions from developers of all experience levels. This guide will help you get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting Pull Requests](#submitting-pull-requests)
- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
  - [Go Style Guide](#go-style-guide)
  - [TypeScript/React Style Guide](#typescriptreact-style-guide)
- [Testing Requirements](#testing-requirements)
- [Commit Message Conventions](#commit-message-conventions)
- [Review Process](#review-process)
- [License Agreement](#license-agreement)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing to ensure a welcoming and inclusive environment for everyone.

## Getting Started

Before you begin:

1. Make sure you have Go 1.22+ and Node.js 18+ installed
2. Familiarize yourself with the project structure and documentation
3. Check existing issues and pull requests to avoid duplicates
4. Join our community discussions for questions and support

## How to Contribute

### Reporting Bugs

Found a bug? Help us improve by reporting it!

**Before submitting a bug report:**

- Check the [existing issues](https://github.com/eth-trading/issues) to avoid duplicates
- Verify the bug exists in the latest version
- Collect relevant information (logs, screenshots, system info)

**Submitting a bug report:**

1. Go to the [Issues page](https://github.com/eth-trading/issues/new)
2. Use the **Bug Report** template
3. Provide a clear, descriptive title
4. Fill in all sections of the template:
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details
   - Logs and error messages
   - Screenshots if applicable

**Bug Report Template:** [.github/ISSUE_TEMPLATE/bug_report.md](.github/ISSUE_TEMPLATE/bug_report.md)

### Suggesting Features

Have an idea to make the project better?

**Before suggesting a feature:**

- Check the [roadmap](IMPLEMENTATION_ROADMAP.md) and [feature gap analysis](FEATURE_GAP_ANALYSIS.md)
- Search existing [feature requests](https://github.com/eth-trading/issues?q=is%3Aissue+label%3Aenhancement)
- Consider whether the feature aligns with project goals

**Submitting a feature request:**

1. Go to the [Issues page](https://github.com/eth-trading/issues/new)
2. Use the **Feature Request** template
3. Provide a clear, descriptive title
4. Explain:
   - The problem or use case
   - Your proposed solution
   - Alternative solutions considered
   - Potential impact on existing functionality
   - Mockups or examples if applicable

**Feature Request Template:** [.github/ISSUE_TEMPLATE/feature_request.md](.github/ISSUE_TEMPLATE/feature_request.md)

### Submitting Pull Requests

Ready to contribute code? Follow these steps:

#### 1. Fork and Clone

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/eth-trading.git
cd eth-trading

# Add upstream remote
git remote add upstream https://github.com/eth-trading/eth-trading.git
```

#### 2. Create a Branch

```bash
# Sync with upstream
git fetch upstream
git checkout master
git merge upstream/master

# Create a feature branch
git checkout -b feat/your-feature-name
# or for bug fixes
git checkout -b fix/bug-description
```

**Branch naming conventions:**

- `feat/feature-name` - New features
- `fix/bug-name` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions or updates
- `chore/description` - Build process or tooling updates

#### 3. Make Your Changes

- Write clean, maintainable code following our [style guidelines](#code-style-guidelines)
- Add tests for new functionality (maintain 80%+ coverage)
- Update documentation as needed
- Keep changes focused and atomic

#### 4. Test Your Changes

```bash
# Run Go tests
make test

# Run Go tests with coverage
make test-coverage

# Run Go linter
make lint

# Format Go code
make fmt

# Run TypeScript/React tests (in web directory)
cd web
npm test

# Run TypeScript linter
npm run lint

# Build the project
cd ..
make build
make web-build
```

#### 5. Commit Your Changes

Follow [Conventional Commits](#commit-message-conventions):

```bash
git add .
git commit -m "feat: add new trading strategy validator"
```

#### 6. Push to Your Fork

```bash
git push origin feat/your-feature-name
```

#### 7. Create a Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Select `master` as the base branch
4. Fill in the PR template with:
   - Clear description of changes
   - Related issue numbers (e.g., "Fixes #123")
   - Testing performed
   - Screenshots for UI changes
   - Breaking changes (if any)

**Pull Request Template:** [.github/PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md)

#### 8. Respond to Review Feedback

- Address reviewer comments promptly
- Make requested changes in new commits
- Re-request review after updates
- Keep discussions respectful and constructive

## Development Setup

### Prerequisites

- **Go**: 1.22 or higher
- **Node.js**: 18.x or higher
- **npm**: 9.x or higher
- **SQLite3**: 3.x or higher
- **Make**: GNU Make 4.x or higher
- **Git**: 2.x or higher

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/eth-trading/eth-trading.git
cd eth-trading

# Install Go dependencies
make deps

# Install web dependencies
make web-install

# Set up environment variables
cp .env.example .env
# Edit .env with your configuration

# Initialize the database
make db-migrate

# Build the project
make build
make web-build
```

### Development Workflow

#### Backend (Go)

```bash
# Run with hot reload (requires air)
make dev

# Or run directly
make run

# Run tests in watch mode
go test ./... -v

# Run specific package tests
go test ./internal/strategy -v
```

#### Frontend (TypeScript/React)

```bash
# Run development server
make web-dev

# Or from web directory
cd web
npm run dev

# Run tests
npm test

# Build for production
npm run build
```

### Project Structure

```
eth-trading/
├── cmd/                    # Application entry points
│   └── bot/               # Main bot application
├── internal/              # Private application code
│   ├── api/              # REST API and WebSocket handlers
│   ├── backtest/         # Backtesting engine
│   ├── binance/          # Binance integration
│   ├── config/           # Configuration management
│   ├── execution/        # Order execution (paper/live)
│   ├── indicators/       # Technical indicators
│   ├── orchestrator/     # Main orchestration logic
│   ├── risk/             # Risk management
│   ├── storage/          # Database and storage
│   └── strategy/         # Trading strategies
├── pkg/                   # Public libraries (if any)
├── web/                   # React frontend
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── hooks/        # Custom hooks
│   │   ├── pages/        # Page components
│   │   ├── services/     # API services
│   │   ├── store/        # State management
│   │   └── types/        # TypeScript types
│   └── public/           # Static assets
├── configs/               # Configuration files
├── data/                  # Runtime data (SQLite, logs)
├── docs/                  # Documentation
└── scripts/               # Build and deployment scripts
```

## Code Style Guidelines

### Go Style Guide

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go) guidelines.

#### Key Principles

**1. Formatting**

```go
// Use gofmt and goimports
make fmt

// Line length: aim for 80-100 characters
// Use tabs for indentation
```

**2. Naming Conventions**

```go
// Packages: lowercase, single word
package strategy

// Interfaces: end with 'er' for single-method interfaces
type Trader interface {
    Execute(order Order) error
}

// Constants: camelCase or MixedCaps
const maxRetries = 3
const DefaultTimeout = 30 * time.Second

// Variables: camelCase
var currentPrice float64
var orderQueue []Order

// Exported names: MixedCaps
type TradingStrategy struct{}
func NewStrategy() *TradingStrategy {}

// Unexported names: mixedCaps
type priceCache struct{}
func calculateProfit() float64 {}
```

**3. Error Handling**

```go
// Always check errors
result, err := executeTrade(order)
if err != nil {
    return fmt.Errorf("failed to execute trade: %w", err)
}

// Use custom error types when appropriate
type ValidationError struct {
    Field string
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}
```

**4. Documentation**

```go
// Package documentation at the top of the file
// Package strategy implements various trading strategies including
// trend following, mean reversion, and volatility-based strategies.
package strategy

// Exported functions and types must have comments
// NewStrategy creates a new trading strategy with the given configuration.
// It returns an error if the configuration is invalid.
func NewStrategy(cfg Config) (*Strategy, error) {
    // implementation
}

// Use full sentences with proper punctuation
// Comments should explain WHY, not WHAT
```

**5. Code Organization**

```go
// Group related declarations
type (
    Order struct {
        ID     string
        Price  float64
        Amount float64
    }

    Position struct {
        Symbol   string
        Quantity float64
    }
)

// Organize imports
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External dependencies
    "github.com/gorilla/websocket"
    "github.com/rs/zerolog/log"

    // Internal packages
    "github.com/eth-trading/internal/config"
    "github.com/eth-trading/internal/types"
)
```

**6. Best Practices**

```go
// Use meaningful variable names
// Good
userCount := len(users)

// Bad
n := len(users)

// Keep functions small and focused
// Prefer early returns
func ProcessOrder(order Order) error {
    if order.Amount <= 0 {
        return ErrInvalidAmount
    }

    if order.Price <= 0 {
        return ErrInvalidPrice
    }

    // Process order
    return nil
}

// Use context for cancellation
func FetchData(ctx context.Context, symbol string) (*Data, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Fetch data
    }
}

// Avoid naked returns except for very short functions
func ValidateOrder(o Order) (valid bool, err error) {
    if o.Amount <= 0 {
        return false, ErrInvalidAmount
    }
    return true, nil
}
```

### TypeScript/React Style Guide

We use TypeScript strict mode and follow the [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript) with React-specific conventions.

#### Key Principles

**1. Formatting**

```typescript
// Use ESLint and Prettier
npm run lint

// Line length: 80-100 characters
// Use 2 spaces for indentation
// Use semicolons
// Use single quotes for strings
```

**2. Naming Conventions**

```typescript
// Components: PascalCase
export const TradingDashboard: React.FC = () => {
  return <div>Dashboard</div>;
};

// Hooks: camelCase with 'use' prefix
export const useTradingData = () => {
  const [data, setData] = useState<TradingData | null>(null);
  return { data, setData };
};

// Types and Interfaces: PascalCase
interface TradingStrategy {
  id: string;
  name: string;
  enabled: boolean;
}

type OrderType = 'MARKET' | 'LIMIT' | 'STOP';

// Constants: UPPER_SNAKE_CASE
const API_BASE_URL = 'http://localhost:8080';
const MAX_RETRY_ATTEMPTS = 3;

// Variables and functions: camelCase
const currentPrice = 1234.56;
const calculateProfit = (entry: number, exit: number): number => {
  return exit - entry;
};
```

**3. TypeScript Types**

```typescript
// Prefer interfaces for object shapes
interface Position {
  symbol: string;
  quantity: number;
  entryPrice: number;
}

// Use type for unions, intersections, and utilities
type OrderStatus = 'PENDING' | 'FILLED' | 'CANCELLED';
type PartialPosition = Partial<Position>;

// Avoid 'any', use 'unknown' when type is truly unknown
const parseJSON = (json: string): unknown => {
  return JSON.parse(json);
};

// Use strict null checks
const findPosition = (id: string): Position | null => {
  // implementation
};

// Explicitly type function parameters and returns
const executeTrade = async (
  order: Order,
  strategy: Strategy
): Promise<TradeResult> => {
  // implementation
};
```

**4. React Components**

```typescript
// Functional components with TypeScript
interface DashboardProps {
  userId: string;
  initialData?: TradingData;
}

export const Dashboard: React.FC<DashboardProps> = ({
  userId,
  initialData
}) => {
  const [data, setData] = useState<TradingData | null>(initialData ?? null);

  useEffect(() => {
    fetchData(userId).then(setData);
  }, [userId]);

  if (!data) {
    return <LoadingSpinner />;
  }

  return (
    <div className="dashboard">
      <h1>Trading Dashboard</h1>
      <DataDisplay data={data} />
    </div>
  );
};

// Prefer named exports over default exports
// Good
export const Button: React.FC<ButtonProps> = (props) => { };

// Avoid
export default function Button(props: ButtonProps) { }
```

**5. Hooks**

```typescript
// Custom hooks for reusable logic
export const useWebSocket = (url: string) => {
  const [data, setData] = useState<Message | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket(url);

    ws.onopen = () => setIsConnected(true);
    ws.onmessage = (event) => setData(JSON.parse(event.data));
    ws.onclose = () => setIsConnected(false);

    return () => ws.close();
  }, [url]);

  return { data, isConnected };
};

// Use dependency arrays correctly
useEffect(() => {
  // Effect code
}, [dep1, dep2]); // Include all dependencies
```

**6. State Management (Zustand)**

```typescript
// Store with TypeScript
interface TradingStore {
  positions: Position[];
  orders: Order[];
  addPosition: (position: Position) => void;
  removePosition: (id: string) => void;
}

export const useTradingStore = create<TradingStore>((set) => ({
  positions: [],
  orders: [],

  addPosition: (position) =>
    set((state) => ({ positions: [...state.positions, position] })),

  removePosition: (id) =>
    set((state) => ({
      positions: state.positions.filter((p) => p.id !== id),
    })),
}));
```

**7. Best Practices**

```typescript
// Use optional chaining and nullish coalescing
const userName = user?.profile?.name ?? 'Guest';

// Destructure props
const Button: React.FC<ButtonProps> = ({
  label,
  onClick,
  disabled = false
}) => {
  return (
    <button onClick={onClick} disabled={disabled}>
      {label}
    </button>
  );
};

// Use async/await over .then()
const fetchData = async (id: string): Promise<Data> => {
  try {
    const response = await api.get(`/data/${id}`);
    return response.data;
  } catch (error) {
    console.error('Failed to fetch data:', error);
    throw error;
  }
};

// Avoid inline styles, use CSS classes
// Good
<div className="trading-card">Content</div>

// Avoid
<div style={{ padding: '10px', margin: '5px' }}>Content</div>
```

## Testing Requirements

We maintain a minimum of **80% code coverage** for all new code. Both unit tests and integration tests are required.

### Go Testing

**Writing Tests**

```go
// File: internal/strategy/mean_reversion_test.go
package strategy

import (
    "testing"
    "github.com/eth-trading/internal/types"
)

func TestMeanReversionStrategy_Generate(t *testing.T) {
    tests := []struct {
        name    string
        prices  []float64
        want    Signal
        wantErr bool
    }{
        {
            name:    "oversold condition",
            prices:  []float64{100, 95, 90, 85, 80},
            want:    SignalBuy,
            wantErr: false,
        },
        {
            name:    "insufficient data",
            prices:  []float64{100},
            want:    SignalNone,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := NewMeanReversionStrategy()
            got, err := s.Generate(tt.prices)

            if (err != nil) != tt.wantErr {
                t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if got != tt.want {
                t.Errorf("Generate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Running Tests**

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/strategy -v

# Run specific test
go test ./internal/strategy -run TestMeanReversionStrategy_Generate -v

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Coverage Requirements**

- New packages: 80%+ coverage
- Bug fixes: Include test case that reproduces the bug
- Critical paths: 90%+ coverage (risk management, execution)
- Integration tests for API endpoints and workflows

### TypeScript/React Testing

**Writing Tests**

```typescript
// File: web/src/components/TradingCard.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { TradingCard } from './TradingCard';

describe('TradingCard', () => {
  it('renders position information correctly', () => {
    const position = {
      symbol: 'ETHUSDT',
      quantity: 1.5,
      entryPrice: 2000,
    };

    render(<TradingCard position={position} />);

    expect(screen.getByText('ETHUSDT')).toBeInTheDocument();
    expect(screen.getByText('1.5')).toBeInTheDocument();
    expect(screen.getByText('$2000')).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', () => {
    const onClose = jest.fn();
    const position = { symbol: 'ETHUSDT', quantity: 1, entryPrice: 2000 };

    render(<TradingCard position={position} onClose={onClose} />);

    fireEvent.click(screen.getByRole('button', { name: /close/i }));

    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
```

**Running Tests**

```bash
# Run all tests
cd web
npm test

# Run with coverage
npm test -- --coverage

# Run in watch mode
npm test -- --watch

# Run specific test file
npm test TradingCard.test.tsx
```

## Commit Message Conventions

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for clear and standardized commit messages.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring without changing functionality
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Build process, tooling, or dependency updates
- `ci`: CI/CD configuration changes
- `revert`: Reverting a previous commit

### Scope (Optional)

The scope specifies what part of the codebase is affected:

- `api`: REST API
- `strategy`: Trading strategies
- `indicators`: Technical indicators
- `risk`: Risk management
- `execution`: Order execution
- `backtest`: Backtesting engine
- `web`: Frontend/UI
- `db`: Database/storage
- `config`: Configuration

### Examples

```bash
# Feature
feat(strategy): add Bollinger Bands breakout strategy

# Bug fix
fix(api): resolve WebSocket connection timeout issue

# Documentation
docs: update README with installation instructions

# Refactoring
refactor(indicators): simplify RSI calculation logic

# Performance
perf(backtest): optimize candle data processing

# Test
test(risk): add unit tests for position sizing

# Chore
chore(deps): upgrade Echo framework to v4.11.4

# Breaking change
feat(api)!: change trade endpoint response format

BREAKING CHANGE: The /api/trades endpoint now returns a different
response structure. Update client code accordingly.
```

### Rules

1. Use the imperative mood ("add" not "added" or "adds")
2. Don't capitalize the first letter of the subject
3. No period at the end of the subject
4. Limit subject line to 72 characters
5. Separate subject from body with a blank line
6. Wrap body at 72 characters
7. Use body to explain what and why, not how
8. Reference issues and pull requests in the footer

### Full Example

```
feat(risk): implement dynamic position sizing

Add Kelly Criterion-based position sizing to optimize risk-adjusted
returns. The algorithm calculates optimal position size based on
win rate, average win/loss ratio, and account balance.

- Add PositionSizer interface
- Implement KellyCriterion calculator
- Add configuration for risk tolerance
- Update risk manager to use dynamic sizing

Closes #45
Related to #38
```

## Review Process

All contributions go through a review process to maintain code quality and project standards.

### What Reviewers Look For

1. **Code Quality**
   - Follows style guidelines
   - Proper error handling
   - Clean, readable code
   - Appropriate abstractions

2. **Testing**
   - Adequate test coverage (80%+)
   - Tests are meaningful and comprehensive
   - Edge cases are covered

3. **Documentation**
   - Code is well-documented
   - README/docs updated if needed
   - API changes documented

4. **Functionality**
   - Code works as intended
   - No regressions introduced
   - Handles edge cases

5. **Performance**
   - No obvious performance issues
   - Efficient algorithms used
   - Resources properly managed

### Review Timeline

- **Initial Review**: Within 2-3 business days
- **Follow-up Reviews**: Within 1-2 business days
- **Final Approval**: Requires at least one maintainer approval

### Responding to Reviews

- Be respectful and constructive
- Address all feedback or explain why you disagree
- Make changes in new commits (don't force push)
- Re-request review after making changes
- Mark conversations as resolved after addressing them

### After Approval

Once approved:

1. A maintainer will merge your PR
2. Your contribution will be included in the next release
3. You'll be added to the contributors list

## License Agreement

By contributing to this project, you agree that your contributions will be licensed under the same license as the project. You certify that:

1. You have the right to submit the contribution
2. You grant the project a perpetual, worldwide, non-exclusive, royalty-free license to use your contribution
3. Your contribution is your original work or you have permission to submit it
4. You understand the contribution may be publicly disclosed

If the project uses a specific license (MIT, Apache, GPL, etc.), contributors must agree to that license's terms.

### Developer Certificate of Origin

By making a contribution, you certify that:

```
Developer Certificate of Origin
Version 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

---

## Questions?

If you have questions about contributing:

- Check existing [documentation](docs/)
- Search [closed issues](https://github.com/eth-trading/issues?q=is%3Aissue+is%3Aclosed)
- Ask in [discussions](https://github.com/eth-trading/discussions)
- Contact the maintainers

## Thank You!

Thank you for taking the time to contribute to ETH Trading Bot. Every contribution, no matter how small, helps make this project better for everyone. We appreciate your effort and look forward to collaborating with you!

Happy coding!
