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

Jan Server uses a multi-layered authentication system supporting multiple authentication methods:

```
┌─────────────────────────────────────────┐
│ Client Application                       │
└──────────────────┬──────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
    ┌───▼────┐          ┌────▼────┐
    │ OAuth2 │          │ API Key  │
    └───┬────┘          └────┬────┘
        │                     │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │ Authentication      │
        │ Gateway             │
        │ (Keycloak/OIDC)    │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │ JWT Validation      │
        │ Token Introspection │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │ Authorization       │
        │ (RBAC/ABAC)        │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │ Protected Resource  │
        └─────────────────────┘
```

### Keycloak/OIDC Integration

#### Configuration

```yaml
# keycloak configuration
keycloak:
  auth_server_url: "https://auth.jan.ai"
  realm: "jan-server"
  client_id: "jan-server"
  client_secret: "${KEYCLOAK_CLIENT_SECRET}"
  
  openid_config:
    issuer: "https://auth.jan.ai/realms/jan-server"
    authorization_endpoint: "https://auth.jan.ai/realms/jan-server/protocol/openid-connect/auth"
    token_endpoint: "https://auth.jan.ai/realms/jan-server/protocol/openid-connect/token"
    userinfo_endpoint: "https://auth.jan.ai/realms/jan-server/protocol/openid-connect/userinfo"
    jwks_uri: "https://auth.jan.ai/realms/jan-server/protocol/openid-connect/certs"
```

#### OAuth2 Code Flow with PKCE

```
User Browser                    Jan Server                  Keycloak
    │                               │                          │
    ├──── Click "Login" ───────────▶│                          │
    │                               │                          │
    │◀─ Redirect to /authorize ────│◀─ Generate code_challenge│
    │ (client_id, state, redirect) │                          │
    │                               │                          │
    ├─────────────────── Call /authorize ──────────────────▶  │
    │                               │                          │
    │◀─ Show login form ─────────────────────────────────────│
    │                               │                          │
    ├──────── Submit credentials ──────────────────────────▶  │
    │                               │                          │
    │◀─ Redirect to callback URL ──────────────────────────│
    │ (code=ABC123, state=xyz)      │                          │
    │                               │                          │
    ├──── Redirect callback ────────▶│                          │
    │                               │                          │
    │                               ├─ Exchange code for token▶
    │                               │ (code, client_id,        │
    │                               │  client_secret,           │
    │                               │  code_verifier)           │
    │                               │                          │
    │                               │◀─ Return JWT token ──────
    │                               │ (access_token,           │
    │                               │  id_token,               │
    │                               │  refresh_token)          │
    │◀─ Set session cookie ────────│                          │
    │                               │                          │
    └──── Redirect to dashboard ────▶                          │
```

#### JWT Token Structure

```
Header:
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key-id-1"
}

Payload:
{
  "iss": "https://auth.jan.ai/realms/jan-server",
  "sub": "user-123",
  "aud": "jan-server",
  "exp": 1672531200,
  "iat": 1672527600,
  "auth_time": 1672527600,
  "name": "John Doe",
  "email": "john@example.com",
  "email_verified": true,
  "realm_access": {
    "roles": ["user", "premium"]
  },
  "resource_access": {
    "jan-server": {
      "roles": ["read:conversations", "write:conversations", "admin"]
    }
  }
}

Signature:
RSASSA-PKCS1-v1_5(
  base64url(header) + '.' + base64url(payload),
  private_key
)
```

#### Token Validation Flow

```python
import jwt
from cryptography.x509 import load_pem_x509_certificate
import requests

class TokenValidator:
    def __init__(self, jwks_uri):
        self.jwks_uri = jwks_uri
        self.public_keys = {}
    
    def get_public_key(self, kid):
        """Fetch public key from JWKS endpoint"""
        if kid not in self.public_keys:
            response = requests.get(self.jwks_uri)
            jwks = response.json()
            
            for key_data in jwks['keys']:
                key_id = key_data['kid']
                # Convert JWK to PEM format
                public_key = self._jwk_to_pem(key_data)
                self.public_keys[key_id] = public_key
        
        return self.public_keys.get(kid)
    
    def validate_token(self, token):
        """Validate JWT token"""
        try:
            # Decode without verification first to get kid
            header = jwt.get_unverified_header(token)
            kid = header['kid']
            
            # Get public key
            public_key = self.get_public_key(kid)
            
            # Verify and decode
            payload = jwt.decode(
                token,
                public_key,
                algorithms=['RS256'],
                audience='jan-server',
                issuer='https://auth.jan.ai/realms/jan-server'
            )
            
            return payload
        
        except jwt.ExpiredSignatureError:
            raise AuthenticationError("Token expired")
        except jwt.InvalidTokenError:
            raise AuthenticationError("Invalid token")
    
    def _jwk_to_pem(self, jwk):
        """Convert JWK to PEM format"""
        # Implementation details omitted
        pass

# Usage in middleware
@app.before_request
def validate_auth():
    token = request.headers.get('Authorization', '').replace('Bearer ', '')
    if token:
        try:
            payload = token_validator.validate_token(token)
            g.user = payload
        except AuthenticationError:
            abort(401)
```

### Guest Login

For unauthenticated users, Jan Server supports guest login:

```
POST /llm/auth/guest-login
├─ No credentials required
├─ Returns temporary token (expires in 1 hour)
├─ Limited to read-only operations
└─ Rate limited: 10 requests per hour per IP
```

---

## Authorization (RBAC)

### Role Hierarchy

```
┌─────────────────────────────────┐
│ Administrator                    │
│ ├─ All permissions              │
│ ├─ Manage users                 │
│ ├─ Manage roles                 │
│ ├─ Access audit logs            │
│ └─ System configuration         │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ Premium User                     │
│ ├─ Create conversations         │
│ ├─ Use all models               │
│ ├─ Webhooks                     │
│ ├─ MCP tools                    │
│ ├─ Batch operations             │
│ └─ Priority support             │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ Standard User                    │
│ ├─ Create conversations         │
│ ├─ Use free models              │
│ ├─ Basic features               │
│ └─ Rate limited                 │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ Guest                            │
│ ├─ Read-only access             │
│ ├─ No conversation creation     │
│ ├─ Limited model access         │
│ └─ Heavily rate limited         │
└─────────────────────────────────┘
```

### Permission Matrix

```
Resource            | Guest | Standard | Premium | Admin
─────────────────────────────────────────────────────
GET /conversations  | NO    | YES      | YES     | YES
POST /conversations | NO    | YES      | YES     | YES
PATCH /conversations| NO    | YES      | YES     | YES
DELETE /conversations| NO    | NO       | YES     | YES
GET /models         | YES   | YES      | YES     | YES
POST /chat/complete | NO    | LIMITED  | YES     | YES
POST /media/upload  | NO    | 5 files  | 100 files| YES
POST /webhooks      | NO    | NO       | YES     | YES
GET /admin/*        | NO    | NO       | NO      | YES
```

### RBAC Implementation

```python
from functools import wraps

def require_role(*roles):
    """Decorator to enforce role-based access"""
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            user_roles = g.user.get('realm_access', {}).get('roles', [])
            if not any(role in user_roles for role in roles):
                abort(403, description="Insufficient permissions")
            return f(*args, **kwargs)
        return decorated_function
    return decorator

def require_permission(resource, action):
    """Decorator to enforce permission-based access"""
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            user_perms = g.user.get('resource_access', {}).get('jan-server', {}).get('roles', [])
            required = f"{action}:{resource}"
            if required not in user_perms:
                abort(403, description="Insufficient permissions")
            return f(*args, **kwargs)
        return decorated_function
    return decorator

# Usage
@app.route('/v1/conversations', methods=['POST'])
@require_role('premium', 'admin')
def create_conversation():
    return {"id": "conv-123"}

@app.route('/v1/webhooks', methods=['POST'])
@require_permission('webhooks', 'write')
def create_webhook():
    return {"id": "hook-123"}
```

---

## Encryption

### Data in Transit (TLS/SSL)

```
TLS 1.3 Configuration:
├─ Cipher suites:
│  ├─ TLS_AES_256_GCM_SHA384 (preferred)
│  ├─ TLS_CHACHA20_POLY1305_SHA256
│  └─ TLS_AES_128_GCM_SHA256
├─ Certificate:
│  ├─ Issued by: DigiCert / Let's Encrypt
│  ├─ Validity: 1 year
│  ├─ Auto-renewal: 30 days before expiry
│  └─ SAN: *.jan.ai, jan.ai
├─ Perfect Forward Secrecy: Enabled
└─ HSTS: max-age=31536000, includeSubDomains
```

#### Certificate Pinning

```python
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.ssl_ import create_urllib3_context

class PinningAdapter(HTTPAdapter):
    def init_poolmanager(self, *args, **kwargs):
        ctx = create_urllib3_context()
        # Add certificate pinning
        ctx.load_verify_locations('server_cert.pem')
        kwargs['ssl_context'] = ctx
        return super().init_poolmanager(*args, **kwargs)

session = requests.Session()
session.mount('https://', PinningAdapter())
response = session.get('https://api.jan.ai/v1/health')
```

### Data at Rest (Database Encryption)

```sql
-- Enable encryption at rest for MySQL
SET GLOBAL innodb_encrypt_tables=ON;
SET GLOBAL innodb_encrypt_log=ON;

-- Encryption key management
SET GLOBAL keyring_encrypted_file_data='/var/lib/mysql-keyring/keyring';
SET GLOBAL keyring_encrypted_file_password='secure-password';

-- Create encrypted table
CREATE TABLE conversations (
  id VARCHAR(36) PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  title VARCHAR(200) NOT NULL ENCRYPTION='Y',
  content LONGTEXT ENCRYPTION='Y',
  created_at TIMESTAMP,
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
) ENCRYPTION='Y';

-- Verify encryption
SELECT TABLE_SCHEMA, TABLE_NAME, CREATE_OPTIONS 
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_NAME='conversations';
```

### Field-Level Encryption

```python
from cryptography.fernet import Fernet
import base64

class FieldEncryption:
    def __init__(self, master_key):
        self.cipher = Fernet(base64.urlsafe_b64encode(master_key.encode()))
    
    def encrypt_field(self, plaintext):
        """Encrypt sensitive field"""
        return self.cipher.encrypt(plaintext.encode()).decode()
    
    def decrypt_field(self, ciphertext):
        """Decrypt sensitive field"""
        return self.cipher.decrypt(ciphertext.encode()).decode()

# Usage
encryption = FieldEncryption(os.getenv('MASTER_ENCRYPTION_KEY'))

# Store
user.email_encrypted = encryption.encrypt_field(user.email)
db.save(user)

# Retrieve
user.email = encryption.decrypt_field(user.email_encrypted)
```

---

## Threat Models & Mitigations

### Threat Matrix

```
┌─────────────────────────────────────────────────────────┐
│ STRIDE Threat Model                                     │
└─────────────────────────────────────────────────────────┘

S - Spoofing Identity
├─ Threat: Attacker impersonates legitimate user
├─ Likelihood: Medium
├─ Impact: High
└─ Mitigation:
   ├─ Mutual TLS authentication
   ├─ Multi-factor authentication (MFA)
   ├─ IP whitelisting for admin
   └─ Hardware security keys

T - Tampering with Data
├─ Threat: Attacker modifies data in transit/at rest
├─ Likelihood: Low
├─ Impact: Critical
└─ Mitigation:
   ├─ TLS 1.3 for transit
   ├─ Database encryption at rest
   ├─ Message authentication codes (HMAC)
   ├─ Cryptographic signatures
   └─ Integrity verification

R - Repudiation
├─ Threat: Attacker denies performing action
├─ Likelihood: Medium
├─ Impact: High
└─ Mitigation:
   ├─ Comprehensive audit logging
   ├─ Non-repudiation tokens
   ├─ Immutable logs
   └─ 3rd-party verification

I - Information Disclosure
├─ Threat: Attacker gains access to sensitive data
├─ Likelihood: Medium
├─ Impact: Critical
└─ Mitigation:
   ├─ Role-based access control (RBAC)
   ├─ Encryption at rest and in transit
   ├─ Data masking in logs
   ├─ Secrets management (Vault)
   └─ DLP (Data Loss Prevention)

D - Denial of Service (DoS)
├─ Threat: Attacker disrupts service availability
├─ Likelihood: High
├─ Impact: High
└─ Mitigation:
   ├─ Rate limiting per user/IP
   ├─ DDoS protection (Cloudflare)
   ├─ Load balancing
   ├─ Circuit breakers
   └─ Auto-scaling

E - Elevation of Privilege
├─ Threat: Attacker gains admin/higher access
├─ Likelihood: Low
├─ Impact: Critical
└─ Mitigation:
   ├─ Principle of least privilege
   ├─ Role-based access control
   ├─ Multi-factor authentication
   ├─ Privilege separation
   └─ Regular access reviews
```

### Attack Scenarios & Responses

#### Scenario 1: Compromised API Token

```
Threat: Attacker obtains user's API token and makes requests

Detection:
├─ Unusual API usage pattern
├─ Request from unfamiliar IP
├─ Rate limit exceeded
└─ Concurrent requests from multiple IPs

Response:
├─ Immediately revoke token
├─ Notify user via email
├─ Require password reset
├─ Enable MFA
├─ Review and audit recent actions
└─ Block suspicious IP addresses
```

#### Scenario 2: SQL Injection

```
Threat: Attacker injects SQL through user input

Prevention:
├─ Parameterized queries (prepared statements)
├─ Input validation and sanitization
├─ ORM usage (SQLAlchemy)
├─ WAF (Web Application Firewall)
└─ Regular security scanning

Code Example:
# UNSAFE
query = f"SELECT * FROM users WHERE email='{email}'"
result = db.execute(query)

# SAFE
query = "SELECT * FROM users WHERE email=?"
result = db.execute(query, (email,))

# SAFER (ORM)
result = User.query.filter_by(email=email).first()
```

#### Scenario 3: Cross-Site Request Forgery (CSRF)

```
Threat: Attacker tricks user into making unwanted request

Prevention:
├─ CSRF tokens on all state-changing operations
├─ SameSite cookie attribute
├─ Verify Origin/Referer headers
└─ User confirmation for sensitive operations

Implementation:
# Generate CSRF token
csrf_token = secrets.token_urlsafe(32)
session['csrf_token'] = csrf_token

# Validate on POST/PUT/PATCH
@app.before_request
def validate_csrf():
    if request.method in ['POST', 'PUT', 'PATCH', 'DELETE']:
        token = request.form.get('_csrf_token', 
                    request.headers.get('X-CSRF-Token'))
        if not token or token != session.get('csrf_token'):
            abort(403, description="CSRF token validation failed")
```

---

## API Security

### Rate Limiting Strategy

```
Tier 1 - Anonymous/Guest
├─ 10 requests per minute per IP
├─ 100 requests per hour per IP
└─ Burst: 20 requests per 10 seconds

Tier 2 - Authenticated User
├─ 100 requests per minute per user
├─ 10,000 requests per hour per user
└─ Burst: 200 requests per 10 seconds

Tier 3 - Premium User
├─ 1,000 requests per minute per user
├─ 1,000,000 requests per hour per user
└─ Burst: 2,000 requests per 10 seconds

Tier 4 - API Key Holder
├─ Configurable per API key
├─ Can request custom limits
└─ Usage tracking and alerts
```

#### Implementation

```python
from flask_limiter import Limiter

limiter = Limiter(
    app,
    key_func=get_client_identifier,
    storage_uri="redis://localhost:6379"
)

@app.route('/v1/chat/completions', methods=['POST'])
@limiter.limit("100/hour;20/minute")
def chat_completion():
    return {"status": "ok"}

def get_client_identifier():
    """Get unique identifier for rate limiting"""
    if g.user:
        return g.user['sub']  # Use user ID
    else:
        return request.remote_addr  # Use IP address
```

### Input Validation

```python
from pydantic import BaseModel, Field, EmailStr

class CreateConversationRequest(BaseModel):
    title: str = Field(..., min_length=1, max_length=200)
    description: str = Field(None, max_length=1000)
    model: str = Field(None, regex=r"^[a-z0-9_-]+$")

@app.route('/v1/conversations', methods=['POST'])
def create_conversation():
    try:
        data = CreateConversationRequest(**request.json)
        # Process validated data
        return {"id": "conv-123"}
    except ValidationError as e:
        return {"errors": e.errors()}, 400
```

### Output Encoding

```python
import html

# Prevent XSS by encoding user input in responses
user_input = '<script>alert("XSS")</script>'
safe_output = html.escape(user_input)
# Result: &lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;

# Use in templates (Jinja2 auto-escapes by default)
# {{ user_input }}  # Automatically escaped
```

---

## Multi-Tenant Isolation

### Data Isolation

```
Tenant A                      Tenant B
├─ Users                      ├─ Users
├─ Conversations              ├─ Conversations
│  ├─ Conv 1                  │  ├─ Conv 5
│  └─ Conv 2                  │  └─ Conv 6
├─ Media Files                ├─ Media Files
│  ├─ File 1                  │  ├─ File 10
│  └─ File 2                  │  └─ File 11
└─ Settings                   └─ Settings

Database Schema:
├─ conversations (tenant_id, user_id, ...)
├─ messages (conversation_id, tenant_id, ...)
├─ media_files (tenant_id, user_id, ...)
└─ audit_logs (tenant_id, user_id, ...)

Query Example:
SELECT * FROM conversations 
WHERE tenant_id = ? AND user_id = ?  -- Always include tenant_id
```

### Tenant Context

```python
from flask import g, request

class TenantContext:
    """Ensure tenant isolation in all operations"""
    
    @staticmethod
    def get_current_tenant():
        """Get tenant ID from JWT token"""
        user = g.user
        return user.get('tenant_id')
    
    @staticmethod
    def validate_tenant_access(resource_tenant_id):
        """Verify user can access this tenant's data"""
        current_tenant = TenantContext.get_current_tenant()
        if resource_tenant_id != current_tenant:
            abort(403, description="Access denied")

@app.route('/v1/conversations/<conversation_id>')
def get_conversation(conversation_id):
    conv = db.get_conversation(conversation_id)
    
    # Always validate tenant
    TenantContext.validate_tenant_access(conv.tenant_id)
    
    return conv.to_dict()
```

### Network Isolation

```yaml
# Kubernetes NetworkPolicy for tenant isolation
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tenant-isolation
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: jan-server
    ports:
    - protocol: TCP
      port: 8000
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: jan-server
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
```

---

## Audit Logging

### Audit Trail

All sensitive operations are logged:

```python
class AuditLog:
    def log(self, event_type, user_id, tenant_id, resource, action, result):
        """Log security event"""
        log_entry = {
            "timestamp": datetime.utcnow().isoformat(),
            "event_type": event_type,
            "user_id": user_id,
            "tenant_id": tenant_id,
            "resource": resource,
            "resource_id": resource.id,
            "action": action,
            "result": result,  # success/failure
            "ip_address": request.remote_addr,
            "user_agent": request.headers.get('User-Agent'),
            "request_id": request.headers.get('X-Request-ID')
        }
        
        # Immutable append
        db.audit_logs.insert(log_entry)
        
        # Also ship to centralized logging
        logger.info(log_entry)

# Usage
@app.route('/v1/conversations', methods=['POST'])
def create_conversation():
    try:
        conv = create_conv()
        AuditLog.log(
            event_type="RESOURCE_CREATE",
            user_id=g.user['sub'],
            tenant_id=g.user['tenant_id'],
            resource=conv,
            action="create",
            result="success"
        )
        return conv.to_dict()
    except Exception as e:
        AuditLog.log(
            event_type="RESOURCE_CREATE",
            user_id=g.user['sub'],
            tenant_id=g.user['tenant_id'],
            resource=None,
            action="create",
            result="failure"
        )
        raise
```

### Audit Events

```
Event Type              | Fields Logged
────────────────────────────────────────────
LOGIN                   | user_id, ip, timestamp, status
TOKEN_GENERATED         | user_id, token_type, expires_at
TOKEN_REVOKED           | user_id, token_id, reason
PASSWORD_CHANGED        | user_id, timestamp, old_hash_required
PERMISSION_GRANTED      | user_id, role, granted_by, timestamp
PERMISSION_REVOKED      | user_id, role, revoked_by, timestamp
CONVERSATION_CREATED    | user_id, conv_id, timestamp
CONVERSATION_DELETED    | user_id, conv_id, timestamp
MEDIA_UPLOADED          | user_id, file_id, size, type
DATA_EXPORTED           | user_id, data_type, records_count
ADMIN_CONFIG_CHANGED    | admin_id, setting, old_value, new_value
FAILED_AUTH_ATTEMPT     | ip, username, reason, timestamp
SUSPICIOUS_ACTIVITY     | user_id, activity_type, details
```

---

## Security Best Practices

### Development

```python
# 1. Use security headers
@app.after_request
def set_security_headers(response):
    response.headers['X-Content-Type-Options'] = 'nosniff'
    response.headers['X-Frame-Options'] = 'DENY'
    response.headers['X-XSS-Protection'] = '1; mode=block'
    response.headers['Strict-Transport-Security'] = 'max-age=31536000'
    response.headers['Content-Security-Policy'] = "default-src 'self'"
    return response

# 2. Secrets management
import os
from dotenv import load_dotenv

load_dotenv()
DB_PASSWORD = os.getenv('DB_PASSWORD')  # Never hardcode
API_KEY = os.getenv('API_KEY')

# 3. Dependency scanning
# pip install safety
# safety check

# 4. Code scanning
# pip install bandit
# bandit -r .

# 5. SAST (Static Application Security Testing)
# Use tools like SonarQube, Snyk

# 6. Regular updates
# pip install --upgrade pip
# pip install -U -r requirements.txt
```

### Deployment

```yaml
# 1. Runtime security
apiVersion: v1
kind: Pod
metadata:
  name: jan-server
spec:
  containers:
  - name: app
    image: jan-server:latest
    securityContext:
      runAsNonRoot: true
      runAsUser: 1000
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL

# 2. Secrets handling
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: db-secret
        key: password

# 3. RBAC
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: jan-server
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
  resourceNames: ["db-secret"]
```

### Monitoring

```
Security Monitoring Checklist:
├─ Failed login attempts
├─ Brute force detection
├─ Unusual API patterns
├─ Permission escalation attempts
├─ Unencrypted data transmission
├─ Certificate expiration
├─ Rate limit violations
├─ Unauthorized access attempts
├─ Configuration changes
├─ Audit log tampering
├─ DLP alerts
└─ Intrusion detection
```

---

## See Also

- [Architecture Overview](./README.md)
- [Performance & SLA Guide](./performance.md)

- [Webhooks Guide](../guides/webhooks.md)

---

**Generated:** December 23, 2025  
**Status:** Production-Ready  
**Version:** v0.0.14
