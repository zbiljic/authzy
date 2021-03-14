package http

// Standard HTTP response constants
const (
	Accept        = "Accept"
	Authorization = "Authorization"
	ContentType   = "Content-Type"
	Cookie        = "Cookie"
	Location      = "Location"
	RetryAfter    = "Retry-After"
	ServerInfo    = "Server"
	SetCookie     = "Set-Cookie"
)

// Non standard HTTP response constants
const (
	XCSRFToken = "X-CSRF-Token"
	XRequestID = "X-Request-ID"
	XUseCookie = "X-Use-Cookie"
)
