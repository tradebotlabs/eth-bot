-- ETH Trading Bot - Rollback Initial Schema Migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_trading_accounts_updated_at ON trading_accounts;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS trading_accounts CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop UUID extension (optional, comment out if other apps use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
