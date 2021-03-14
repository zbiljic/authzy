package api

import "time"

// SignupRequest are the parameters the signup endpoint accepts.
type SignupRequest struct {
	Email        string                 `json:"email,omitempty"`
	Password     string                 `json:"password,omitempty"`
	Username     string                 `json:"username,omitempty"`
	GivenName    string                 `json:"given_name,omitempty"`
	FamilyName   string                 `json:"family_name,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Nickname     string                 `json:"nickname,omitempty"`
	Picture      string                 `json:"picture,omitempty"`
	UserMetaData map[string]interface{} `json:"user_metadata,omitempty"`
}

type SignupResponse struct {
	ID            string `json:"user_id"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Username      string `json:"username,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	FamilyName    string `json:"family_name,omitempty"`
	Name          string `json:"name,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	Picture       string `json:"picture,omitempty"`
}

type UserUpdateRequest struct {
	Email            string                 `json:"email"`
	EmailChangeToken string                 `json:"email_change_token"`
	Password         string                 `json:"password"`
	Username         string                 `json:"username"`
	GivenName        string                 `json:"given_name"`
	FamilyName       string                 `json:"family_name"`
	Name             string                 `json:"name"`
	Nickname         string                 `json:"nickname"`
	Picture          string                 `json:"picture"`
	AppMetaData      map[string]interface{} `json:"app_metadata"`
	UserMetaData     map[string]interface{} `json:"user_metadata"`
}

type UserResponse struct {
	UserID        string                 `json:"user_id"`
	Email         string                 `json:"email,omitempty"`
	EmailVerified bool                   `json:"email_verified,omitempty"`
	Username      string                 `json:"username,omitempty"`
	GivenName     string                 `json:"given_name,omitempty"`
	FamilyName    string                 `json:"family_name,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Nickname      string                 `json:"nickname,omitempty"`
	Picture       string                 `json:"picture,omitempty"`
	AppMetaData   map[string]interface{} `json:"app_metadata,omitempty"`
	UserMetaData  map[string]interface{} `json:"user_metadata,omitempty"`
	Providers     []Provider             `json:"providers,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type Provider struct {
	Provider    string `json:"provider"`
	FederatedID string `json:"federated_id"`
}

// VerifyRequest are the parameters the verify endpoint accepts.
type VerifyRequest struct {
	Type       string `json:"type"`
	Token      string `json:"token"`
	Password   string `json:"password"`
	RedirectTo string `json:"redirect_to"`
}

// RecoverRequest holds the parameters for a password recovery request.
type RecoverRequest struct {
	Email string `json:"email"`
}

// CustomClaims is a struct thats used for JWT claims.
type CustomClaims struct {
	Username     string                 `json:"username,omitempty"`
	Email        string                 `json:"email,omitempty"`
	AppMetaData  map[string]interface{} `json:"app_metadata,omitempty"`
	UserMetaData map[string]interface{} `json:"user_metadata,omitempty"`
}

// AccessTokenResponse represents an OAuth2 success response
type AccessTokenResponse struct {
	Token        string `json:"access_token"`
	TokenType    string `json:"token_type"` // Bearer
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Introspection contains an access token's session data as specified
// by IETF RFC 7662, see: https://tools.ietf.org/html/rfc7662
type Introspection struct {
	// Active is a boolean indicator of whether or not the presented token
	// is currently active.  The specifics of a token's "active" state
	// will vary depending on the implementation of the authorization
	// server and the information it keeps about its tokens, but a "true"
	// value return for the "active" property will generally indicate
	// that a given token has been issued by this authorization server,
	// has not been revoked by the resource owner, and is within its
	// given time window of validity (e.g., after its issuance time and
	// before its expiration time).
	//
	// required: true
	Active bool `json:"active"`

	// Scope is a JSON string containing a space-separated list of
	// scopes associated with this token.
	Scope string `json:"scope,omitempty"`

	// ID is a client identifier for the OAuth 2.0 client that
	// requested this token.
	ClientID string `json:"client_id"`

	// Subject of the token, as defined in JWT [RFC7519].
	// Usually a machine-readable identifier of the resource owner who
	// authorized this token.
	Subject string `json:"sub"`

	// ObfuscatedSubject is set when the subject identifier algorithm was set to "pairwise" during authorization.
	// It is the `sub` value of the ID Token that was issued.
	ObfuscatedSubject string `json:"obfuscated_subject,omitempty"`

	// Expires at is an integer timestamp, measured in the number of seconds
	// since January 1 1970 UTC, indicating when this token will expire.
	ExpiresAt int64 `json:"exp"`

	// Issued at is an integer timestamp, measured in the number of seconds
	// since January 1 1970 UTC, indicating when this token was
	// originally issued.
	IssuedAt int64 `json:"iat"`

	// NotBefore is an integer timestamp, measured in the number of seconds
	// since January 1 1970 UTC, indicating when this token is not to be
	// used before.
	NotBefore int64 `json:"nbf"`

	// Username is a human-readable identifier for the resource owner who
	// authorized this token.
	Username string `json:"username,omitempty"`

	// Audience contains a list of the token's intended audiences.
	Audience []string `json:"aud"`

	// IssuerURL is a string representing the issuer of this token
	Issuer string `json:"iss"`

	// TokenType is the introspected token's type, typically `Bearer`.
	TokenType string `json:"token_type"`

	// TokenUse is the introspected token's use, for example `access_token` or `refresh_token`.
	TokenUse string `json:"token_use"`

	// Extra is arbitrary data set by the session.
	Extra map[string]interface{} `json:"ext,omitempty"`
}

// UserinfoResponse The userinfo response.
//
// see: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
type UserinfoResponse struct {
	Sub                 string               `json:"sub,omitempty"`
	Name                string               `json:"name,omitempty"`
	GivenName           string               `json:"given_name,omitempty"`
	FamilyName          string               `json:"family_name,omitempty"`
	MiddleName          string               `json:"middle_name,omitempty"`
	Nickname            string               `json:"nickname,omitempty"`
	PreferredUsername   string               `json:"preferred_username,omitempty"`
	Profile             string               `json:"profile,omitempty"`
	Picture             string               `json:"picture,omitempty"`
	Website             string               `json:"website,omitempty"`
	Email               string               `json:"email,omitempty"`
	EmailVerified       bool                 `json:"email_verified,omitempty"`
	Gender              string               `json:"gender,omitempty"`
	Birthdate           string               `json:"birthdate,omitempty"`
	Zoneinfo            string               `json:"zoneinfo,omitempty"`
	Locale              string               `json:"locale,omitempty"`
	PhoneNumber         string               `json:"phone_number,omitempty"`
	PhoneNumberVerified bool                 `json:"phone_number_verified,omitempty"`
	Address             UserinfoAddressClaim `json:"address,omitempty"`
	UpdatedAt           int64                `json:"updated_at,omitempty"`
}

// UserinfoAddressClaim represents a physical mailing address.
//
// see: https://openid.net/specs/openid-connect-core-1_0.html#AddressClaim
type UserinfoAddressClaim struct {
	Formatted     string `json:"formatted,omitempty"`
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}
