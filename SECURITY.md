# Security Policy

## Supported Versions

We take security seriously. The following versions of ETH Trading Bot are currently being supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

We appreciate your efforts to responsibly disclose your findings and will make every effort to acknowledge your contributions.

### How to Report

To report a security vulnerability, please use one of the following methods:

#### 1. GitHub Security Advisories (Preferred)

1. Navigate to the repository's [Security Advisories](https://github.com/yourusername/eth-trading/security/advisories)
2. Click "Report a vulnerability"
3. Fill out the advisory details form
4. Submit the report

#### 2. Email

Send an email to: **security@eth-trading.dev**

Include the following information:
- Type of vulnerability
- Full paths of source file(s) related to the manifestation of the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Varies based on severity and complexity

### What to Expect

After submitting a vulnerability report:

1. **Acknowledgment**: We'll confirm receipt within 48 hours
2. **Assessment**: We'll investigate and assess the severity
3. **Communication**: We'll keep you informed of our progress
4. **Resolution**: We'll develop and test a fix
5. **Disclosure**: We'll coordinate a responsible disclosure timeline
6. **Credit**: We'll publicly acknowledge your contribution (unless you prefer to remain anonymous)

## Security Best Practices

### For Users

If you're deploying ETH Trading Bot, please follow these security best practices:

#### API Key Security

- **Never commit API keys** to version control
- **Use environment variables** or secure configuration management
- **Enable IP whitelisting** on your exchange account
- **Disable withdrawal permissions** on trading API keys
- **Use read-only keys** for market data when possible
- **Rotate API keys** regularly (every 90 days recommended)

#### Database Security

- **Change default passwords** immediately after setup
- **Use strong, unique passwords** for database access
- **Enable SSL/TLS** for database connections in production
- **Restrict database access** to localhost or VPN only
- **Backup encryption keys** securely
- **Regular backups** with encryption

#### Authentication

- **Change default JWT secret** in production (use `openssl rand -base64 32`)
- **Use HTTPS** for all API communications
- **Enable secure cookie flags** (HttpOnly, Secure, SameSite)
- **Implement rate limiting** on authentication endpoints
- **Monitor failed login attempts**
- **Consider enabling 2FA** for admin accounts

#### Network Security

- **Use firewall rules** to restrict access
- **Deploy behind reverse proxy** (nginx/Caddy) with HTTPS
- **Implement rate limiting** to prevent abuse
- **Monitor for unusual traffic patterns**
- **Use VPN** for remote access to trading systems

#### Application Security

- **Keep dependencies updated** (check monthly for security patches)
- **Run with least privileges** (don't use root user)
- **Enable audit logging** for security events
- **Monitor system resources** for anomalies
- **Test updates** in staging environment first

#### Secure Configuration

```yaml
# Production config.yaml example
postgres:
  host: "localhost"  # Never expose to public internet
  sslmode: "require"  # Always use SSL in production

auth:
  jwtSecret: "USE_OPENSSL_RAND_BASE64_32_TO_GENERATE"  # CHANGE THIS!
  tokenExpiry: 15m
  refreshTokenExpiry: 168h

api:
  port: ":8080"
  corsOrigins:
    - "https://yourdomain.com"  # Only allow your frontend domain
```

### For Developers

#### Secure Coding Practices

- **Validate all inputs** from users and external APIs
- **Sanitize SQL queries** (use parameterized queries - already implemented with sqlx)
- **Implement rate limiting** on all endpoints
- **Use bcrypt** for password hashing (cost factor 12+)
- **Never log sensitive data** (passwords, API keys, tokens)
- **Implement proper error handling** (don't expose stack traces to users)

#### Code Review Checklist

Before submitting PRs, verify:

- [ ] No hardcoded credentials or API keys
- [ ] Input validation on all user-provided data
- [ ] Proper error handling without information leakage
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (proper output encoding)
- [ ] Authentication/authorization checks on protected endpoints
- [ ] No sensitive data in logs
- [ ] Dependencies are up to date

#### Dependency Management

- **Regularly update dependencies** for security patches
- **Review new dependencies** before adding them
- **Use `go mod verify`** to check module integrity
- **Monitor security advisories** for Go and npm packages

```bash
# Check for outdated Go packages
go list -u -m all

# Check for npm vulnerabilities
cd web && npm audit

# Update dependencies (after testing!)
go get -u ./...
cd web && npm update
```

## Known Security Considerations

### Current Limitations

1. **API Keys Storage**: Binance secret keys should be encrypted at rest (planned for v2.0)
2. **Email Verification**: Not yet implemented (planned for v2.0)
3. **Two-Factor Authentication**: Not yet implemented (planned for v2.0)
4. **Rate Limiting**: Should be implemented at reverse proxy level
5. **CAPTCHA**: Not implemented for registration/login

### Planned Security Enhancements

- [ ] API key encryption at rest (AES-256)
- [ ] Email verification system
- [ ] Two-factor authentication (TOTP)
- [ ] Account lockout after failed login attempts
- [ ] IP-based access controls
- [ ] Webhook signature verification
- [ ] Security headers middleware
- [ ] CAPTCHA for authentication endpoints

## Security Audit History

| Date | Auditor | Scope | Status |
|------|---------|-------|--------|
| TBD  | Community Review | Full codebase | Planned for v1.0 release |

## Disclosure Policy

When we receive a security bug report, we will:

1. **Confirm the problem** and determine affected versions
2. **Audit code** to find similar problems
3. **Prepare fixes** for all supported versions
4. **Release patched versions** as soon as possible

We aim for responsible disclosure:

- **Embargo Period**: 90 days or until a fix is released, whichever comes first
- **Public Disclosure**: After fixes are released, we'll publish a security advisory
- **Credit**: Security researchers will be credited (unless anonymous reporting is requested)

## Security Contacts

- **Security Email**: security@eth-trading.dev
- **GPG Key**: Available at [keybase.io/ethtrading](https://keybase.io/ethtrading)
- **GitHub Security**: https://github.com/yourusername/eth-trading/security

## Security Hall of Fame

We're grateful to the following security researchers who have helped improve ETH Trading Bot:

*No reports yet - be the first!*

---

## Disclaimer

This trading bot handles potentially sensitive financial data and API credentials. While we implement security best practices, **no software is 100% secure**.

- Always use strong, unique passwords
- Never expose your trading bot directly to the internet
- Monitor your accounts regularly
- Start with small amounts and demo accounts
- You are responsible for securing your deployment

**Use at your own risk. The maintainers are not responsible for any financial losses or security breaches.**

---

*Last updated: 2025-12-26*
