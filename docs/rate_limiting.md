# Rate Limiting Documentation

## Overview

The BucketBall Backend implements a professional, multi-layered rate limiting system to prevent brute force attacks and ensure system stability. The rate limiting is specifically designed for login attempts and provides both email-based and IP-based protection.

## Features

### 🔒 **Dual-Layer Protection**
- **Email-based rate limiting**: Tracks failed login attempts per email address
- **IP-based rate limiting**: Tracks failed login attempts per IP address
- **Both layers must pass** for login to be allowed

### ⚡ **Smart Rate Limiting**
- **5 failed attempts** within **15 minutes** triggers a lockout
- **30-minute lockout** period after exceeding the limit
- **Automatic reset** on successful login
- **Separate counters** for email and IP addresses

### 🛡️ **Security Features**
- **Redis-based storage** for fast, distributed rate limiting
- **Automatic cleanup** of expired rate limit data
- **Admin tools** for monitoring and resetting rate limits
- **Detailed logging** of rate limit events

## Configuration

### Default Settings
```go
MaxAttempts:     5           // Maximum failed attempts allowed
WindowDuration:  15 minutes  // Time window for counting attempts
LockoutDuration: 30 minutes  // Lockout period after exceeding limit
```

### Customization
Rate limiting settings can be modified in `services/rate_limit_service.go`:

```go
var DefaultLoginRateLimit = RateLimitConfig{
    MaxAttempts:     5,                    // Adjust as needed
    WindowDuration:  15 * time.Minute,     // Adjust time window
    LockoutDuration: 30 * time.Minute,     // Adjust lockout duration
}
```

## API Endpoints

### Login Endpoint
```http
POST /api/v1/auth/login
```

**Rate Limiting Behavior:**
- ✅ **Success**: Clears all rate limit counters
- ❌ **Failed**: Increments rate limit counters
- 🚫 **Rate Limited**: Returns `429 Too Many Requests`

**Response Examples:**

**Successful Login:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": { ... }
  }
}
```

**Rate Limited:**
```json
{
  "success": false,
  "message": "too many login attempts. Please try again in 25m0s",
  "error": "Rate limit exceeded"
}
```

### Admin Endpoints

#### Get Rate Limit Information
```http
GET /api/v1/admin/rate-limit/info?email=user@example.com&ip=192.168.1.1
```

**Response:**
```json
{
  "success": true,
  "message": "Rate limit info retrieved",
  "data": {
    "email_attempts": 3,
    "ip_attempts": 1,
    "email_locked": false,
    "ip_locked": false,
    "email_lockout_remaining": "0s",
    "ip_lockout_remaining": "0s"
  }
}
```

#### Reset Rate Limits
```http
POST /api/v1/admin/rate-limit/reset?email=user@example.com&ip=192.168.1.1
```

**Response:**
```json
{
  "success": true,
  "message": "Rate limits reset for both email and IP"
}
```

## Implementation Details

### Rate Limit Keys
- **Email-based**: `login_attempts:email:{email}`
- **IP-based**: `login_attempts:ip:{ip}`
- **Lockout keys**: `lockout:login_attempts:email:{email}`

### Redis Data Structure
```redis
# Attempt counter (expires after WindowDuration)
login_attempts:email:user@example.com -> 3
login_attempts:ip:192.168.1.1 -> 1

# Lockout flag (expires after LockoutDuration)
lockout:login_attempts:email:user@example.com -> 1
lockout:login_attempts:ip:192.168.1.1 -> 1
```

### Flow Diagram
```
Login Request
     ↓
Check Email Rate Limit
     ↓
Check IP Rate Limit
     ↓
Both Allowed? → No → Return 429
     ↓ Yes
Attempt Login
     ↓
Success? → Yes → Clear Counters → Return Token
     ↓ No
Increment Counters
     ↓
Exceed Limit? → Yes → Set Lockout → Return 401
     ↓ No
Return 401
```

## Security Considerations

### ✅ **Protection Against**
- Brute force attacks
- Credential stuffing
- Distributed attacks (per IP)
- Account enumeration

### ⚠️ **Important Notes**
- Rate limits are **per email AND per IP**
- Successful login **clears all counters**
- Lockout periods are **cumulative** (both email and IP must be clear)
- Admin can **reset rate limits** for legitimate users

### 🔧 **Monitoring**
- Use admin endpoints to monitor rate limit status
- Check Redis keys directly for debugging
- Monitor failed login patterns
- Set up alerts for repeated lockouts

## Testing

### Test Rate Limiting
```bash
# Test normal login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"wrongpassword"}'

# Repeat 5 times to trigger rate limit
# Should return 429 after 5th attempt
```

### Test Admin Reset
```bash
# Check rate limit status
curl -X GET "http://localhost:8080/api/v1/admin/rate-limit/info?email=test@example.com" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# Reset rate limits
curl -X POST "http://localhost:8080/api/v1/admin/rate-limit/reset?email=test@example.com" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## Troubleshooting

### Common Issues

1. **Rate limit not working**
   - Check Redis connection
   - Verify Redis keys are being set
   - Check client IP detection

2. **False positives**
   - Use admin reset endpoint
   - Check for shared IP addresses
   - Verify email normalization

3. **Performance issues**
   - Monitor Redis memory usage
   - Check Redis connection pool
   - Consider Redis clustering for high load

### Debug Commands
```bash
# Check Redis keys
redis-cli keys "login_attempts:*"
redis-cli keys "lockout:*"

# Check specific user
redis-cli get "login_attempts:email:user@example.com"
redis-cli ttl "lockout:login_attempts:email:user@example.com"
```

## Best Practices

1. **Monitor rate limit metrics** regularly
2. **Set up alerts** for unusual patterns
3. **Document rate limit policies** for users
4. **Test rate limiting** in staging environment
5. **Have admin tools** ready for legitimate resets
6. **Consider user experience** when setting limits
7. **Log rate limit events** for security analysis
