# Security Architecture Deep Dive

> **Status:** v0.0.14 | **Last Updated:** December 23, 2025 | **Level:** Advanced

Comprehensive security documentation covering authentication, authorization, encryption, threat models, and defense strategies.

## Table of Contents

- [Authentication Architecture](#authentication-architecture)
- [Authorization (RBAC)](#authorization-rbac)
- [Encryption](#encryption)
- [Threat Models & Mitigations](#threat-models--mitigations)
- [API Security](#api-security)
- [Multi-Tenant Isolation](#multi-tenant-isolation)
- [Audit Logging](#audit-logging)
- [Security Best Practices](#security-best-practices)

---

## Authentication Architecture

### Overview

Jan Server uses a multi-layered authentication system supporting multiple authentication methods: OAuth2 and API Key authentication flow through an Authentication Gateway (Keycloak/OIDC), followed by JWT Validation and Token Introspection, then Authorization (RBAC/ABAC), and finally access to Protected Resources.

### Keycloak/OIDC Integration

#### Configuration

Configure Keycloak with auth_server_url, realm, client_id, client_secret, and openid_config endpoints (issuer, authorization_endpoint, token_endpoint, userinfo_endpoint, jwks_uri).

#### OAuth2 Code Flow with PKCE

The OAuth2 flow with PKCE follows this sequence: User clicks login → Jan Server generates code_challenge and redirects to Keycloak /authorize → User submits credentials → Keycloak redirects to callback with authorization code → Jan Server exchanges code for JWT tokens (access_token, id_token, refresh_token) using code_verifier → Session cookie is set and user is redirected to dashboard.

#### JWT Token Structure

JWT tokens contain three parts: Header (algorithm RS256, type JWT, key ID), Payload (issuer, subject, audience, expiration, issued at, auth time, user details, roles and permissions), and Signature (RSASSA-PKCS1-v1_5 signature using private key).

#### Token Validation Flow

Token validation involves: fetching public keys from JWKS endpoint, decoding the JWT header to extract the key ID, retrieving the corresponding public key, verifying the signature using RS256 algorithm, validating issuer and audience claims, checking expiration, and returning the payload. Middleware validates tokens on each request.

### Guest Login

For unauthenticated users, Jan Server supports guest login via POST /llm/auth/guest-login endpoint. No credentials are required. Returns a temporary token (expires in 1 hour). Limited to read-only operations and rate limited to 10 requests per hour per IP.

---

## Authorization (RBAC)

### Role Hierarchy

**Administrator**: All permissions, manage users and roles, access audit logs, system configuration.

**Premium User**: Create conversations, use all models, webhooks, MCP tools, batch operations, priority support.

**Standard User**: Create conversations, use free models, basic features, rate limited.

**Guest**: Read-only access, no conversation creation, limited model access, heavily rate limited.

### Permission Matrix

Permissions vary by role:
- **Guest**: Can only GET /models
- **Standard**: Can GET/POST/PATCH conversations, GET models, LIMITED chat completions, upload 5 files
- **Premium**: All Standard permissions plus DELETE conversations, unlimited chat completions, upload 100 files, webhooks
- **Admin**: All permissions including admin endpoints

### RBAC Implementation

RBAC is implemented using decorators that check user roles and permissions from JWT tokens. Role-based decorators verify if a user has required roles (e.g., premium, admin). Permission-based decorators check specific resource:action permissions (e.g., write:webhooks). Both abort with 403 if requirements are not met.

---

## Encryption

### Data in Transit (TLS/SSL)

**TLS 1.3 Configuration**: Uses cipher suites TLS_AES_256_GCM_SHA384 (preferred), TLS_CHACHA20_POLY1305_SHA256, and TLS_AES_128_GCM_SHA256. Certificates issued by DigiCert/Let's Encrypt with 1-year validity and auto-renewal 30 days before expiry. SAN covers *.jan.ai and jan.ai. Perfect Forward Secrecy enabled. HSTS configured with max-age=31536000 and includeSubDomains.

#### Certificate Pinning

Certificate pinning is implemented using custom HTTP adapters that create an SSL context and load specific server certificates for verification. This prevents man-in-the-middle attacks by ensuring only trusted certificates are accepted.

### Data at Rest (Database Encryption)

Database encryption at rest uses MySQL InnoDB encryption features: enable innodb_encrypt_tables and innodb_encrypt_log, configure keyring for key management, create tables with ENCRYPTION='Y' option for sensitive columns (title, content). Verify encryption status by querying INFORMATION_SCHEMA.TABLES.

### Field-Level Encryption

Field-level encryption uses Fernet symmetric encryption with a master key. Sensitive fields are encrypted before storage and decrypted when retrieved. The master key is stored securely in environment variables.

---

## Threat Models & Mitigations

### Threat Matrix

**STRIDE Threat Model**:

**S - Spoofing Identity**: Attacker impersonates legitimate user (Medium likelihood, High impact). Mitigations: Mutual TLS, MFA, IP whitelisting, hardware security keys.

**T - Tampering with Data**: Attacker modifies data in transit/at rest (Low likelihood, Critical impact). Mitigations: TLS 1.3, database encryption, HMAC, cryptographic signatures, integrity verification.

**R - Repudiation**: Attacker denies performing action (Medium likelihood, High impact). Mitigations: Comprehensive audit logging, non-repudiation tokens, immutable logs, 3rd-party verification.

**I - Information Disclosure**: Attacker gains access to sensitive data (Medium likelihood, Critical impact). Mitigations: RBAC, encryption at rest and in transit, data masking, secrets management (Vault), DLP.

**D - Denial of Service**: Attacker disrupts service availability (High likelihood, High impact). Mitigations: Rate limiting, DDoS protection (Cloudflare), load balancing, circuit breakers, auto-scaling.

**E - Elevation of Privilege**: Attacker gains admin/higher access (Low likelihood, Critical impact). Mitigations: Principle of least privilege, RBAC, MFA, privilege separation, regular access reviews.

### Attack Scenarios & Responses

#### Scenario 1: Compromised API Token

**Threat**: Attacker obtains user's API token and makes requests.

**Detection**: Unusual API usage pattern, requests from unfamiliar IPs, rate limit exceeded, concurrent requests from multiple IPs.

**Response**: Immediately revoke token, notify user via email, require password reset, enable MFA, review and audit recent actions, block suspicious IP addresses.

#### Scenario 2: SQL Injection

**Threat**: Attacker injects SQL through user input.

**Prevention**: Use parameterized queries (prepared statements), input validation and sanitization, ORMs (SQLAlchemy), WAF (Web Application Firewall), regular security scanning. Always use parameterized queries or ORM methods instead of string concatenation.

#### Scenario 3: Cross-Site Request Forgery (CSRF)

**Threat**: Attacker tricks user into making unwanted request.

**Prevention**: Use CSRF tokens on all state-changing operations, SameSite cookie attribute, verify Origin/Referer headers, user confirmation for sensitive operations. Generate unique CSRF tokens per session and validate them on all POST/PUT/PATCH/DELETE requests.

---

## API Security

### Rate Limiting Strategy

**Tier 1 - Anonymous/Guest**: 10 requests/minute, 100 requests/hour per IP, burst 20 requests/10 seconds.

**Tier 2 - Authenticated User**: 100 requests/minute, 10,000 requests/hour per user, burst 200 requests/10 seconds.

**Tier 3 - Premium User**: 1,000 requests/minute, 1,000,000 requests/hour per user, burst 2,000 requests/10 seconds.

**Tier 4 - API Key Holder**: Configurable per API key, can request custom limits, usage tracking and alerts.

#### Implementation

Rate limiting is implemented using decorators with Redis storage backend. The key function determines whether to rate limit by user ID (authenticated) or IP address (anonymous). Limits are specified per endpoint using decorator syntax.

### Input Validation

Input validation uses schema-based validation with Pydantic models. Define field constraints (min_length, max_length, regex patterns) and validate request data before processing. Return 400 errors with detailed validation messages on failure.

### Output Encoding

Prevent XSS attacks by HTML-encoding user input in responses. Use html.escape() for manual encoding or rely on template engine auto-escaping (Jinja2 auto-escapes by default).

---

## Multi-Tenant Isolation

### Data Isolation

Each tenant has isolated data: users, conversations, media files, and settings. Database schema includes tenant_id in all tables (conversations, messages, media_files, audit_logs). All queries must include tenant_id filter to ensure data isolation.

### Tenant Context

Tenant context is extracted from JWT tokens and validated on every resource access. A TenantContext class provides methods to get the current tenant ID and validate that users can only access resources belonging to their tenant, aborting with 403 on violations.

### Network Isolation

Kubernetes NetworkPolicy enforces tenant isolation at the network level. Policies restrict ingress traffic to port 8000 from jan-server namespace only, and egress traffic to PostgreSQL (port 5432) within the same namespace.

---

## Audit Logging

### Audit Trail

All sensitive operations are logged with an AuditLog class that captures: timestamp, event_type, user_id, tenant_id, resource, action, result (success/failure), ip_address, user_agent, and request_id. Log entries are immutably appended to the database and shipped to centralized logging. Every endpoint logs both successful and failed operations.

### Audit Events

Audit events include: LOGIN (user_id, ip, status), TOKEN_GENERATED/REVOKED (user_id, token details), PASSWORD_CHANGED (user_id, timestamp), PERMISSION_GRANTED/REVOKED (user_id, role, actor), CONVERSATION_CREATED/DELETED (user_id, conv_id), MEDIA_UPLOADED (user_id, file details), DATA_EXPORTED (user_id, data_type, count), ADMIN_CONFIG_CHANGED (admin_id, setting, values), FAILED_AUTH_ATTEMPT (ip, username, reason), SUSPICIOUS_ACTIVITY (user_id, activity details).

---

## Security Best Practices

### Development

**Security Headers**: Set X-Content-Type-Options (nosniff), X-Frame-Options (DENY), X-XSS-Protection (1; mode=block), Strict-Transport-Security (max-age=31536000), Content-Security-Policy (default-src 'self') on all responses.

**Secrets Management**: Load secrets from environment variables using dotenv, never hardcode credentials.

**Dependency Scanning**: Use safety to check for vulnerable dependencies.

**Code Scanning**: Use bandit for Python security linting.

**SAST**: Use tools like SonarQube and Snyk for static analysis.

**Regular Updates**: Keep dependencies up to date with pip upgrade.

### Deployment

**Runtime Security**: Configure pod securityContext with runAsNonRoot: true, runAsUser: 1000, readOnlyRootFilesystem: true, allowPrivilegeEscalation: false, drop all capabilities.

**Secrets Handling**: Load secrets from Kubernetes secretKeyRef instead of environment variables.

**RBAC**: Define minimal Kubernetes roles that only allow access to specific secrets with get verb.

### Monitoring

**Security Monitoring Checklist**: Failed login attempts, brute force detection, unusual API patterns, permission escalation attempts, unencrypted data transmission, certificate expiration, rate limit violations, unauthorized access attempts, configuration changes, audit log tampering, DLP alerts, intrusion detection.

---

## See Also

- [Architecture Overview](./README.md)
- [Performance & SLA Guide](./performance.md)

- [Webhooks Guide](../guides/webhooks.md)

---

**Generated:** December 23, 2025  
**Status:** Production-Ready  
**Version:** v0.0.14
