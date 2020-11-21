package token

import "context"

const authTokenClaimsCtxPath = "AUTH_TOKEN_CLAIMS"

// FromContext retreives the AuthToken from gRPC request
// context for use in the rpc method. If the token is not
// found in context, or is malformed, nil will be returned
// instead.
func FromContext(ctx context.Context) *AuthToken {
	val := ctx.Value(authTokenClaimsCtxPath)
	if val == nil {
		return nil
	}

	token, ok := val.(AuthToken)
	if !ok {
		return nil
	}

	return &token
}

// WithAuthToken returns the provided context with an AuthToken
// attached. This is useful for sharing user information across
// rpc methods/middlewares.
func WithAuthToken(ctx context.Context, token AuthToken) context.Context {
	return context.WithValue(ctx, authTokenClaimsCtxPath, token)
}
