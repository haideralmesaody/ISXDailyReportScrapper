# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in ISX Pulse, please report it responsibly.

**DO NOT** create a public GitHub issue for security vulnerabilities.

Instead, please email the maintainers with:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if available)

We will acknowledge receipt within 48 hours and provide a detailed response within 7 days.

## Security Standards

This project implements multiple security layers:

### 1. License Protection
- Hardware-locked licensing system
- AES-256-GCM encryption
- Secure key derivation with scrypt
- Tamper detection mechanisms

### 2. Credential Security
- Encrypted credential storage
- Hardware fingerprint-based decryption
- No plaintext secrets in code or logs
- Secure credential rotation support

### 3. API Security
- JWT authentication with short-lived tokens
- Rate limiting on all endpoints
- Input validation and sanitization
- CSRF protection
- XSS prevention

### 4. Data Protection
- Encryption at rest for sensitive data
- Secure communication channels
- Audit logging for security events
- No logging of sensitive information

## Security Best Practices

### For Contributors

1. **Never commit secrets**:
   - Use `.env` files (never commit them)
   - Use `encrypted_credentials.dat`
   - Check commits for accidental secrets

2. **Validate all inputs**:
   - Sanitize user inputs
   - Use parameterized queries
   - Implement proper bounds checking

3. **Follow secure coding**:
   - Use crypto/rand for randomness
   - Implement proper error handling
   - Avoid race conditions
   - Use context timeouts

4. **Test security**:
   - Write security-focused tests
   - Test authentication flows
   - Verify authorization checks
   - Test rate limiting

### For Users

1. **Protect credentials**:
   - Keep `credentials.json` secure
   - Use strong passwords
   - Rotate credentials regularly
   - Monitor for unauthorized access

2. **Secure deployment**:
   - Use HTTPS in production
   - Keep software updated
   - Implement firewall rules
   - Monitor security logs

3. **License security**:
   - Protect `license.dat` file
   - Don't share license keys
   - Report suspicious activity

## Security Checklist

Before each release:

- [ ] Run security scanner (e.g., gosec)
- [ ] Check for vulnerable dependencies
- [ ] Review authentication code
- [ ] Verify encryption implementations
- [ ] Test rate limiting
- [ ] Audit logging functionality
- [ ] Review error messages (no info leakage)
- [ ] Validate CORS configuration
- [ ] Check CSP headers
- [ ] Test input validation

## Known Security Considerations

### WebSocket Security
- Implements rate limiting
- Validates message formats
- Handles connection drops gracefully
- Prevents message flooding

### File Operations
- Validates file paths
- Prevents directory traversal
- Implements file size limits
- Uses atomic operations

### Frontend Security
- Implements CSP headers
- Sanitizes user inputs
- Prevents XSS attacks
- Handles hydration securely

## Dependencies

Regular dependency audits are performed:

### Go Dependencies
```bash
go list -json -m all | nancy sleuth
```

### Node Dependencies
```bash
cd dev/frontend && npm audit
```

## Compliance

This project aims to comply with:
- OWASP Top 10 guidelines
- OWASP ASVS standards
- Industry best practices for financial data

## Security Updates

Security patches are prioritized and released as soon as possible.

Monitor:
- GitHub Security Advisories
- Release notes for security updates
- CHANGELOG.md for security fixes

## Contact

For security concerns, contact the maintainers directly.

For general security questions, consult:
- `docs/SECURITY.md` (detailed implementation)
- `.claude/agents/security-auditor.md` (security review)
- CLAUDE.md (security standards)

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security researchers who help improve the project's security.