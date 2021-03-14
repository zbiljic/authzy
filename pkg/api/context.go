package api

import (
	"context"
	"net"

	"github.com/lestrrat-go/jwx/jwt"

	"github.com/zbiljic/authzy/pkg/config"
)

type contextKey string

func (c contextKey) String() string {
	return "api context key " + string(c)
}

const (
	configKey    = contextKey("config")
	requestIDKey = contextKey("request_id")
	tokenKey     = contextKey("jwt")
	userIPKey    = contextKey("user_ip")
)

// withConfig adds the configuration to the context.
func withConfig(ctx context.Context, config *config.Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}

func getConfig(ctx context.Context) *config.Config {
	if obj := ctx.Value(configKey); obj != nil {
		return obj.(*config.Config)
	}
	return nil
}

// withRequestID adds the provided request ID to the context.
func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// getRequestID reads the request ID from the context.
func getRequestID(ctx context.Context) string {
	if obj := ctx.Value(requestIDKey); obj != nil {
		return obj.(string)
	}
	return ""
}

// withToken adds the JWT token to the context.
func withToken(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// getToken reads the JWT token from the context.
func getToken(ctx context.Context) *jwt.Token {
	if obj := ctx.Value(tokenKey); obj != nil {
		return obj.(*jwt.Token)
	}
	return nil
}

// withUserIP adds the user IP address to the context.
func withUserIP(ctx context.Context, userIP net.IP) context.Context {
	return context.WithValue(ctx, userIPKey, userIP)
}

// getUserIP reads the user IP address from the context.
func getUserIP(ctx context.Context) net.IP {
	if obj := ctx.Value(userIPKey); obj != nil {
		return obj.(net.IP)
	}
	return nil
}
