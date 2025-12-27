-- ETH Trading Bot - Development Seed Data
-- WARNING: This file contains test data only. DO NOT use in production!

-- Clean existing data (careful!)
TRUNCATE TABLE audit_logs, sessions, trading_accounts, users RESTART IDENTITY CASCADE;

-- Insert test users
-- Password for admin user: "admin"
-- Hash generated with bcrypt cost 12
INSERT INTO users (id, email, password_hash, full_name, role, is_active, is_email_verified, created_at, updated_at) VALUES
(
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    'admin@gmail.com',
    '$2a$12$8z0HOpsJN8xPrCZKBm0zDOVHMyP5s/nGgRJXxPxhQ6w6qJLwTZfH.',  -- admin
    'Admin User',
    'admin',
    true,
    true,
    NOW(),
    NOW()
),
(
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    'trader@eth-trading.local',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/vQqZPqrHqX8s7eR0O',  -- TestPassword123
    'Demo Trader',
    'trader',
    true,
    true,
    NOW(),
    NOW()
),
(
    '550e8400-e29b-41d4-a716-446655440003'::uuid,
    'viewer@eth-trading.local',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/vQqZPqrHqX8s7eR0O',  -- TestPassword123
    'Viewer User',
    'viewer',
    true,
    true,
    NOW(),
    NOW()
),
(
    '550e8400-e29b-41d4-a716-446655440004'::uuid,
    'live-trader@eth-trading.local',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/vQqZPqrHqX8s7eR0O',  -- TestPassword123
    'Live Trader',
    'trader',
    true,
    true,
    NOW(),
    NOW()
);

-- Insert test trading accounts
INSERT INTO trading_accounts (
    id,
    user_id,
    account_type,
    account_name,
    demo_initial_capital,
    demo_current_balance,
    trading_symbol,
    trading_mode,
    enabled_strategies,
    is_active,
    created_at,
    updated_at
) VALUES
(
    '660e8400-e29b-41d4-a716-446655440001'::uuid,
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    'demo',
    'Admin Demo Account',
    50000.00,
    50000.00,
    'ETHUSDT',
    'paper',
    ARRAY['TrendFollowing', 'MeanReversion'],
    true,
    NOW(),
    NOW()
),
(
    '660e8400-e29b-41d4-a716-446655440002'::uuid,
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    'demo',
    'Small Capital Test',
    5000.00,
    5000.00,
    'ETHUSDT',
    'paper',
    ARRAY['TrendFollowing', 'Breakout'],
    true,
    NOW(),
    NOW()
),
(
    '660e8400-e29b-41d4-a716-446655440003'::uuid,
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    'demo',
    'Large Capital Test',
    100000.00,
    100000.00,
    'ETHUSDT',
    'paper',
    ARRAY['TrendFollowing', 'MeanReversion', 'Breakout', 'Volatility', 'StatArb'],
    true,
    NOW(),
    NOW()
),
(
    '660e8400-e29b-41d4-a716-446655440004'::uuid,
    '550e8400-e29b-41d4-a716-446655440004'::uuid,
    'live',
    'Live Trading Account (Testnet)',
    NULL,
    NULL,
    'ETHUSDT',
    'paper',  -- Start with paper mode even for live accounts
    ARRAY['TrendFollowing'],
    true,
    NOW(),
    NOW()
);

-- Update live account with dummy API credentials (for testing only)
UPDATE trading_accounts
SET
    binance_api_key = 'test_api_key_do_not_use_in_production',
    binance_api_key_masked = 'test_...ction',
    binance_testnet = true
WHERE id = '660e8400-e29b-41d4-a716-446655440004'::uuid;

-- Insert sample audit log entries
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, ip_address, created_at) VALUES
(
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    'user.login',
    'user',
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    '127.0.0.1',
    NOW() - INTERVAL '1 hour'
),
(
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    'account.created',
    'trading_account',
    '660e8400-e29b-41d4-a716-446655440002'::uuid,
    '127.0.0.1',
    NOW() - INTERVAL '2 days'
);

-- Display seed data summary
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SEED DATA LOADED SUCCESSFULLY';
    RAISE NOTICE '========================================';
    RAISE NOTICE '';
    RAISE NOTICE 'Test Users:';
    RAISE NOTICE '  - admin@gmail.com (Admin) - Password: admin';
    RAISE NOTICE '  - trader@eth-trading.local (Trader) - Password: TestPassword123';
    RAISE NOTICE '  - viewer@eth-trading.local (Viewer) - Password: TestPassword123';
    RAISE NOTICE '  - live-trader@eth-trading.local (Live) - Password: TestPassword123';
    RAISE NOTICE '';
    RAISE NOTICE '';
    RAISE NOTICE 'Trading Accounts:';
    RAISE NOTICE '  - Admin Demo Account ($50,000)';
    RAISE NOTICE '  - Small Capital Test ($5,000)';
    RAISE NOTICE '  - Large Capital Test ($100,000)';
    RAISE NOTICE '  - Live Trading Account (Testnet)';
    RAISE NOTICE '';
    RAISE NOTICE 'WARNING: This is test data only!';
    RAISE NOTICE 'DO NOT use these credentials in production.';
    RAISE NOTICE '========================================';
    RAISE NOTICE '';
END $$;

-- Verify data
SELECT
    'Users' AS table_name,
    COUNT(*)::text AS count,
    string_agg(email, ', ') AS details
FROM users
UNION ALL
SELECT
    'Trading Accounts' AS table_name,
    COUNT(*)::text AS count,
    string_agg(account_name, ', ') AS details
FROM trading_accounts
UNION ALL
SELECT
    'Sessions' AS table_name,
    COUNT(*)::text AS count,
    'N/A' AS details
FROM sessions
UNION ALL
SELECT
    'Audit Logs' AS table_name,
    COUNT(*)::text AS count,
    'N/A' AS details
FROM audit_logs;
