# User Signup API

## Endpoint
`POST /api/v1/auth/register`

## Description
Register a new user account with comprehensive profile information including personal details, location, and financial fields.

## Request Body

### Required Fields
- `email` (string): User's email address (must be valid email format)
- `username` (string): Unique username (3-20 characters)
- `password` (string): User's password (minimum 6 characters)
- `first_name` (string): User's first name (2-50 characters)
- `last_name` (string): User's last name (2-50 characters)

### Optional Fields
- `profile_pic` (string): URL or path to user's profile picture
- `dob` (string): Date of birth in YYYY-MM-DD format
- `phone_number` (string): Phone number (10-15 characters)
- `location` (object): User's location information
  - `country` (string): Country name (2-50 characters)
  - `state` (string): State/Province name (2-50 characters)
  - `city` (string): City name (2-50 characters)
  - `postal_code` (string): Postal/ZIP code (3-10 characters)

## Example Request

```json
{
  "email": "john.doe@example.com",
  "username": "johndoe",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe",
  "profile_pic": "https://example.com/avatars/john.jpg",
  "dob": "1990-05-15",
  "phone_number": "+1234567890",
  "location": {
    "country": "United States",
    "state": "California",
    "city": "San Francisco",
    "postal_code": "94102"
  }
}
```

## Response

### Success Response (201 Created)
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "email": "john.doe@example.com",
      "username": "johndoe",
      "first_name": "John",
      "last_name": "Doe",
      "profile_pic": "https://example.com/avatars/john.jpg",
      "dob": "1990-05-15T00:00:00Z",
      "phone_number": "+1234567890",
      "location": {
        "country": "United States",
        "state": "California",
        "city": "San Francisco",
        "postal_code": "94102"
      },
      "balance": 0.0,
      "withdraw": 0.0,
      "role": "user",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

### Error Responses

#### 400 Bad Request - Validation Error
```json
{
  "success": false,
  "message": "Validation failed",
  "error": "Field validation errors..."
}
```

#### 409 Conflict - User Already Exists
```json
{
  "success": false,
  "message": "User with this email already exists",
  "error": null
}
```

#### 409 Conflict - Username Taken
```json
{
  "success": false,
  "message": "Username already taken",
  "error": null
}
```

## Validation Rules

### Email
- Must be a valid email format
- Required field

### Username
- Must be 3-20 characters long
- Required field
- Must be unique

### Password
- Minimum 6 characters
- Required field

### First Name
- Must be 2-50 characters long
- Required field

### Last Name
- Must be 2-50 characters long
- Required field

### Profile Picture
- Optional field
- Should be a valid URL or file path

### Date of Birth
- Optional field
- Must be in YYYY-MM-DD format if provided

### Phone Number
- Optional field
- Must be 10-15 characters if provided

### Location
- Optional object
- All location fields are optional but validated if provided
- Country: 2-50 characters
- State: 2-50 characters
- City: 2-50 characters
- Postal Code: 3-10 characters

## Rate Limiting
- 5 requests per minute per IP address

## Security Features
- Password is hashed using bcrypt
- JWT token is generated for authentication
- Token is stored in Redis for session management
- Input validation on all fields
- Rate limiting to prevent abuse

## Notes
- All new users start with a balance of 0.0
- All new users start with withdraw amount of 0.0
- Users are created with "user" role by default
- Users are created as active by default
- CreatedAt and UpdatedAt timestamps are automatically set
