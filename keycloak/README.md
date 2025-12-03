# Keycloak Setup for Jan Server

This directory contains Keycloak configuration for Jan Server's user management, authentication, and feature flag system.

## Overview

Jan Server uses Keycloak as the **single source of truth** for:
- User authentication and identity management
- Role-Based Access Control (RBAC)
- Group membership and organization
- Feature flags (stored as group attributes)

**Architecture**: JWT-only approach - all user data, groups, roles, and feature flags are included in JWT claims. No server-side caching or database sync needed.

## Directory Structure

```
keycloak/
├── README.md              # This file - setup documentation
├── import/
│   └── realm-jan.json     # Realm configuration for auto-import
└── init/
    └── (initialization scripts if needed)
```

## Quick Start

### 1. Start Keycloak with Docker Compose

Keycloak is included in the Jan Server docker-compose setup:

```bash
# Start full stack including Keycloak
docker-compose up -d keycloak

# Or start everything
docker-compose up -d
```

Keycloak will automatically import the realm configuration from `import/realm-jan.json` on first startup.

### 2. Access Keycloak Admin Console

- **URL**: http://localhost:8080 (or configured port)
- **Admin Console**: http://localhost:8080/admin
- **Default Admin Credentials**: 
  - Username: `admin`
  - Password: `admin` (⚠️ Change in production!)

### 3. Verify Realm Import

1. Log into Admin Console
2. Select **jan** realm from dropdown (top-left)
3. Verify:
   - ✅ Clients: `backend`, `jan-client`
   - ✅ Roles: `admin`, `user`, `guest`
   - ✅ Groups: `jan_group`, `pilot_users`, `standard`, `guest`
   - ✅ Group attributes: `feature_flags` configured

## Realm Configuration

### Realm: jan

- **Realm Name**: `jan`
- **Registration**: Enabled
- **Email as Username**: Yes
- **Email Verification**: Disabled (for development)
- **Reset Password**: Disabled
- **Token Exchange**: Enabled

### Clients

#### 1. backend (Service Account)
- **Client ID**: `backend`
- **Type**: Confidential (has secret)
- **Purpose**: Backend service-to-service authentication
- **Secret**: `backend-secret` (⚠️ Change in production!)
- **Grants**: Service Account, Direct Access
- **Token Lifespan**: 300 seconds (5 minutes)
- **Use Case**: Admin API operations, Keycloak management

#### 2. jan-client (Public Client)
- **Client ID**: `jan-client`
- **Type**: Public (no secret)
- **Purpose**: User authentication for web/mobile apps
- **Grants**: Authorization Code Flow with PKCE
- **Token Lifespan**: 
  - Access Token: 3600 seconds (1 hour)
  - Refresh Token: 2592000 seconds (30 days)
- **Redirect URIs**: Localhost + production domains
- **Protocol Mappers**: Groups, feature flags, email_verified, realm roles

### Roles

#### Realm Roles

| Role | Description | Capabilities |
|------|-------------|--------------|
| `admin` | Administrator | Full system access, manage users/groups/flags |
| `user` | Registered User | Standard platform access |
| `guest` | Guest User | Temporary/limited access |

**Admin Detection**: Check for `admin` in `realm_access.roles` array in JWT

### Groups

Groups are used for organization and feature flag assignment. All groups have `feature_flags` attribute.

| Group | Path | Auto-Assign | Feature Flags | Purpose |
|-------|------|-------------|---------------|---------|
| `jan_group` | `/jan_group` | Auto (@jan.ai, @menlo.ai emails) | `experimental_models` | Internal team members |
| `pilot_users` | `/pilot_users` | Manual | `experimental_models` | Beta testers, early adopters |
| `standard` | `/standard` | Auto (verified email) | None | Regular verified users |
| `guest` | `/guest` | Auto (guest login) | None | Guest/temporary access |

**Default Group**: New users are automatically added to `/standard` group

### Feature Flags

Feature flags are stored as Keycloak group attributes and included in JWT tokens.

#### Current Feature Flags

| Flag Key | Description | Groups with Access |
|----------|-------------|-------------------|
| `experimental_models` | Access to experimental/beta models in model catalog | `jan_group`, `pilot_users` |

#### Adding Feature Flags to Groups

**Via Admin Console:**
1. Navigate to **Groups** → Select group
2. Go to **Attributes** tab
3. Add attribute:
   - Key: `feature_flags`
   - Value: `["experimental_models"]` (JSON array)
4. Click **Add** then **Save**

**Via Admin API (Go):**
```go
import "github.com/Nerzal/gocloak/v13"

// Add feature flag to group
group, err := client.GetGroup(ctx, token, realm, groupID)
if err != nil {
    return err
}

if group.Attributes == nil {
    group.Attributes = &map[string][]string{}
}

(*group.Attributes)["feature_flags"] = []string{"experimental_models"}

err = client.UpdateGroup(ctx, token, realm, *group)
```

## Protocol Mappers

Protocol mappers add custom claims to JWT tokens.

### jan-client Mappers

#### 1. Groups Mapper
- **Type**: Group Membership
- **Claim Name**: `groups`
- **Full Path**: Yes (includes `/` prefix)
- **Included In**: ID token, Access token, Userinfo
- **Example**: `["jan_group", "/pilot_users"]`

#### 2. Feature Flags Mapper
- **Type**: Group Membership (configured for attributes)
- **Claim Name**: `feature_flags`
- **Source**: Aggregated from all group `feature_flags` attributes
- **Included In**: ID token, Access token, Userinfo
- **Example**: `["experimental_models"]`

#### 3. Realm Roles Mapper
- **Type**: User Realm Role
- **Claim Name**: `realm_access.roles`
- **Multivalued**: Yes
- **Included In**: ID token, Access token
- **Example**: `{"realm_access": {"roles": ["admin", "user"]}}`

#### 4. Email Verified Mapper
- **Type**: User Property
- **Claim Name**: `email_verified`
- **JSON Type**: Boolean
- **Included In**: ID token, Access token
- **Example**: `"email_verified": true`

#### 5. Additional Standard Mappers
- `preferred_username` - Username claim
- `guest` - Guest user attribute
- `pid` - Process/session ID attribute

## JWT Token Structure

### Example Access Token Claims

```json
{
  "exp": 1701648000,
  "iat": 1701644400,
  "jti": "token-uuid",
  "iss": "http://localhost:8080/realms/jan",
  "aud": "jan-client",
  "sub": "user-uuid",
  "typ": "Bearer",
  "azp": "jan-client",
  "session_state": "session-uuid",
  "acr": "1",
  "realm_access": {
    "roles": ["admin", "user"]
  },
  "scope": "openid profile email",
  "sid": "session-uuid",
  "email_verified": true,
  "name": "John Doe",
  "preferred_username": "john.doe",
  "given_name": "John",
  "family_name": "Doe",
  "email": "john.doe@jan.ai",
  "groups": ["/jan_group", "/standard"],
  "feature_flags": ["experimental_models"]
}
```

### Important Claims for Jan Server

| Claim | Type | Purpose | Example |
|-------|------|---------|---------|
| `sub` | string | Unique user ID | `"a1b2c3d4-..."` |
| `email` | string | User email | `"user@jan.ai"` |
| `email_verified` | boolean | Email verification status | `true` |
| `name` | string | Full name | `"John Doe"` |
| `preferred_username` | string | Username | `"john.doe"` |
| `groups` | array | Group paths with `/` prefix | `["/jan_group"]` |
| `realm_access.roles` | array | Realm roles | `["admin", "user"]` |
| `feature_flags` | array | Enabled feature flags | `["experimental_models"]` |

## User Management

### Creating Users

#### Via Admin Console
1. Select **jan** realm
2. Navigate to **Users** → **Add user**
3. Fill in details:
   - Username (will be email)
   - Email (required)
   - Email Verified (check if trusted)
   - Enabled (check to activate)
4. Click **Create**
5. Go to **Credentials** tab → Set password
6. Go to **Groups** tab → Assign groups
7. Go to **Role Mappings** tab → Assign realm roles

#### Via Admin API
See `pkg/keycloak/users.go` for Go client implementation.

### Assigning Admin Role

**Option 1: Realm Role (Recommended)**
1. Navigate to **Users** → Select user
2. Go to **Role Mappings** tab
3. Select **admin** from Available Roles
4. Click **Add selected**

**Option 2: User Attribute (Fallback)**
1. Navigate to **Users** → Select user
2. Go to **Attributes** tab
3. Add attribute: `is_admin` = `true`

### Managing Group Membership

1. Navigate to **Users** → Select user
2. Go to **Groups** tab
3. Select groups from Available Groups
4. Click **Join**

**Auto-Assignment** (requires custom event listener):
- Users with `@jan.ai` or `@menlo.ai` emails → `jan_group`
- Verified users → `standard`
- Guest login → `guest`

## Security Best Practices

### Production Configuration

#### 1. Change Default Credentials
```bash
# Set via environment variables
KEYCLOAK_ADMIN=your-admin-username
KEYCLOAK_ADMIN_PASSWORD=your-secure-password
```

#### 2. Update Client Secrets
- Generate strong random secrets for confidential clients
- Store in secrets management system (Vault, AWS Secrets Manager, etc.)
- Update `docker-compose.yml` and application config

#### 3. Configure HTTPS
```yaml
# In docker-compose.yml
keycloak:
  environment:
    KC_HOSTNAME: keycloak.yourdomain.com
    KC_HOSTNAME_STRICT: true
    KC_HOSTNAME_STRICT_HTTPS: true
    KC_PROXY: edge
```

#### 4. Enable Email Verification
```json
// In realm-jan.json
{
  "verifyEmail": true,
  "loginWithEmailAllowed": true
}
```

Configure SMTP settings in Keycloak Admin Console → Realm Settings → Email.

#### 5. Enable MFA for Admins
1. Navigate to **Authentication** → **Required Actions**
2. Enable **Configure OTP**
3. Set as default for admin users

#### 6. Enable Audit Logging
```json
// In realm-jan.json
{
  "eventsEnabled": true,
  "eventsExpiration": 7776000,
  "eventsListeners": ["jboss-logging"],
  "adminEventsEnabled": true,
  "adminEventsDetailsEnabled": true
}
```

### Token Security

#### Token Lifespans (Production Recommendations)
- **Access Token**: 15-30 minutes (not 1 hour)
- **Refresh Token**: 30 days with rotation
- **SSO Session**: 12 hours
- **Offline Token**: 90 days (if needed)

#### Token Validation
Jan Server validates tokens using JWKS endpoint:
```
http://localhost:8080/realms/jan/protocol/openid-connect/certs
```

Configure in `config/defaults.yaml`:
```yaml
auth:
  jwks_url: "http://keycloak:8080/realms/jan/protocol/openid-connect/certs"
  issuer: "http://keycloak:8080/realms/jan"
  audience: "jan-client"
```

## Troubleshooting

### Common Issues

#### 1. Realm Not Imported
**Symptom**: "jan" realm not found

**Solution**:
```bash
# Check Keycloak logs
docker logs jan-server-keycloak-1

# Manually import realm
docker exec -it jan-server-keycloak-1 /opt/keycloak/bin/kc.sh import \
  --file /opt/keycloak/data/import/realm-jan.json
```

#### 2. Feature Flags Not in JWT
**Symptom**: `feature_flags` claim missing

**Solution**:
1. Verify group attributes are set correctly (JSON array format)
2. Check protocol mapper configuration
3. Issue new token (refresh or re-login)
4. Verify mapper is assigned to client

#### 3. Admin Role Not Working
**Symptom**: Admin endpoints return 403 Forbidden

**Solution**:
1. Check `realm_access.roles` in JWT (use jwt.io)
2. Verify user has `admin` realm role assigned
3. Check middleware is looking for correct role name
4. Try fallback `is_admin` attribute

#### 4. Group Paths Mismatch
**Symptom**: Group detection fails

**Solution**:
Groups have leading slash in JWT: `/jan_group` not `jan_group`
```go
// Correct comparison
if contains(groups, "/jan_group") { ... }

// Or strip prefix
groupName := strings.TrimPrefix(group, "/")
```

#### 5. Token Validation Fails
**Symptom**: 401 Unauthorized errors

**Solution**:
1. Verify JWKS URL is accessible from Jan Server
2. Check issuer and audience match
3. Ensure clock sync between services
4. Check token expiration times

### Debug JWT Contents

Use jwt.io or command line:
```bash
# Decode JWT (requires jq)
echo $ACCESS_TOKEN | cut -d. -f2 | base64 -d | jq

# Or use online tool
# https://jwt.io
```

### Keycloak Logs

```bash
# View logs
docker logs -f jan-server-keycloak-1

# Increase log level
docker exec jan-server-keycloak-1 /opt/keycloak/bin/kc.sh start-dev --log-level=DEBUG
```

## Development vs Production

### Development Settings (Current)
- ✅ Auto-registration enabled
- ✅ Email verification disabled
- ✅ HTTP allowed
- ✅ Localhost redirect URIs
- ⚠️ Weak secrets
- ⚠️ Simple passwords allowed

### Production Checklist
- [ ] Disable auto-registration (or add CAPTCHA)
- [ ] Enable email verification
- [ ] Enforce HTTPS only
- [ ] Configure production redirect URIs
- [ ] Strong secrets in vault
- [ ] Strong password policy
- [ ] Enable MFA for admins
- [ ] Configure SMTP for emails
- [ ] Set up database persistence (PostgreSQL)
- [ ] Enable audit logging
- [ ] Set appropriate token lifespans
- [ ] Configure rate limiting
- [ ] Set up monitoring/alerts

## Integration with Jan Server

### Middleware Configuration

Jan Server uses JWT middleware to extract claims:

```go
// services/llm-api/internal/middleware/auth.go

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Validate JWT (Kong or middleware)
        claims := extractJWTClaims(c)
        
        // Store in context
        c.Set("claims", claims)
        c.Set("user_id", claims["sub"])
        c.Set("user_email", claims["email"])
        c.Set("user_groups", claims["groups"])
        c.Set("feature_flags", claims["feature_flags"])
        
        c.Next()
    }
}
```

### Feature Flag Resolution

```go
// Check if user has feature flag
func IsFeatureEnabled(c *gin.Context, flagKey string) bool {
    claims := c.MustGet("claims").(map[string]interface{})
    
    featureFlagsRaw, ok := claims["feature_flags"]
    if !ok {
        return false
    }
    
    switch flags := featureFlagsRaw.(type) {
    case []interface{}:
        for _, flag := range flags {
            if flagStr, ok := flag.(string); ok && flagStr == flagKey {
                return true
            }
        }
    case []string:
        for _, flag := range flags {
            if flag == flagKey {
                return true
            }
        }
    }
    
    return false
}
```

### Admin Operations

Admin operations use Keycloak Admin API via gocloak client:

```go
// pkg/keycloak/client.go

import "github.com/Nerzal/gocloak/v13"

type KeycloakClient struct {
    client      *gocloak.GoCloak
    realm       string
    clientID    string
    clientSecret string
}

// Example: List users
func (k *KeycloakClient) ListUsers(ctx context.Context) ([]*gocloak.User, error) {
    token, err := k.getAdminToken(ctx)
    if err != nil {
        return nil, err
    }
    
    users, err := k.client.GetUsers(ctx, token.AccessToken, k.realm, gocloak.GetUsersParams{})
    return users, err
}
```

## API Endpoints

### Token Endpoints

```bash
# Get token (password grant)
curl -X POST http://localhost:8080/realms/jan/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=jan-client" \
  -d "username=user@jan.ai" \
  -d "password=password"

# Refresh token
curl -X POST http://localhost:8080/realms/jan/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "client_id=jan-client" \
  -d "refresh_token=$REFRESH_TOKEN"

# Get JWKS (public keys)
curl http://localhost:8080/realms/jan/protocol/openid-connect/certs
```

### Admin API Examples

```bash
# Get admin token
ADMIN_TOKEN=$(curl -X POST http://localhost:8080/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" \
  -d "username=admin" \
  -d "password=admin" | jq -r .access_token)

# List users
curl -X GET http://localhost:8080/admin/realms/jan/users \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Get user groups
curl -X GET http://localhost:8080/admin/realms/jan/users/{user-id}/groups \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Add user to group
curl -X PUT http://localhost:8080/admin/realms/jan/users/{user-id}/groups/{group-id} \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## References

- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [Keycloak Admin REST API](https://www.keycloak.org/docs-api/latest/rest-api/index.html)
- [gocloak Go Client](https://github.com/Nerzal/gocloak)
- [Jan Server User Management Guide](../docs/guides/user-management-todo.md)
- [JWT Specification](https://datatracker.ietf.org/doc/html/rfc7519)

## Support

For issues related to:
- **Keycloak Setup**: Check Keycloak logs and documentation
- **Jan Server Integration**: See `docs/guides/user-management-todo.md`
- **Feature Flags**: Review group attributes and protocol mappers
- **JWT Issues**: Use jwt.io to decode and verify claims

---

**Last Updated**: December 3, 2025
**Keycloak Version**: 23.x (or latest stable)
**Jan Server Version**: See main README.md
