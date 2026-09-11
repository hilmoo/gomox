package msession

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ory/herodot"
)

// sessionDuration is how long a newly created session remains valid.
const sessionDuration = 7 * 24 * time.Hour

// NewSessionParams describes a session row to persist, as built by [CreateSession].
type NewSessionParams struct {
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	IPAddress *string
	UserAgent *string
}

// CreateSessionParams configures [CreateSession].
type CreateSessionParams struct {
	// Secret is used to hash the generated session token before it is persisted (see
	// [HashSessionToken]); the same secret must be supplied to [New] for the session
	// to later validate.
	Secret string
	// UserID is the user the session belongs to.
	UserID uuid.UUID
	// IPAddress and UserAgent are optionally recorded alongside the session.
	IPAddress *string
	UserAgent *string
	// CreateSession persists a session row, mapping any storage error into a herodot
	// error (e.g. via [github.com/hilmoo/gomox/sqlx.MapPgxErrorHTTP]).
	CreateSession func(ctx context.Context, params NewSessionParams) *herodot.DefaultError
}

// CreateSession generates a new session token for args.UserID, persists it via
// args.CreateSession, and returns the raw (unhashed) token for the caller to set as a
// cookie via [SetNewCookies].
func CreateSession(ctx context.Context, args CreateSessionParams) (string, *herodot.DefaultError) {
	token, err := generateRandomString()
	if err != nil {
		return "", herodot.ErrInternalServerError.WithReason("failed to generate session token").WithDebug(err.Error())
	}

	hashToken := HashSessionToken(args.Secret, token)

	if hErr := args.CreateSession(ctx, NewSessionParams{
		UserID:    args.UserID,
		Token:     hashToken,
		ExpiresAt: time.Now().Add(sessionDuration),
		IPAddress: args.IPAddress,
		UserAgent: args.UserAgent,
	}); hErr != nil {
		return "", hErr
	}

	return token, nil
}
