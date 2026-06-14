package livekit

import (
	"errors"
	"fmt"
	"time"

	"github.com/livekit/protocol/auth"
)

const (
	RolePublisher  = "publisher"
	RoleSubscriber = "subscriber"
)

var (
	ErrNotConfigured = errors.New("livekit is not configured")
	ErrInvalidRole   = errors.New("role must be publisher or subscriber")
)

type Config struct {
	URL       string
	APIKey    string
	APISecret string
	TokenTTL  time.Duration
}

func (config Config) Enabled() bool {
	return config.URL != "" && config.APIKey != "" && config.APISecret != ""
}

type TokenRequest struct {
	DeviceID   string
	Identity   string
	Role       string
	ValidFor   time.Duration
}

type TokenResponse struct {
	Token     string
	URL       string
	Room      string
	Identity  string
	ExpiresAt time.Time
}

type TokenIssuer struct {
	config Config
}

func NewTokenIssuer(config Config) *TokenIssuer {
	return &TokenIssuer{config: config}
}

func (issuer *TokenIssuer) Enabled() bool {
	return issuer.config.Enabled()
}

func (issuer *TokenIssuer) IssueToken(request TokenRequest) (TokenResponse, error) {
	if !issuer.config.Enabled() {
		return TokenResponse{}, ErrNotConfigured
	}

	role := request.Role
	if role == "" {
		role = RoleSubscriber
	}
	if role != RolePublisher && role != RoleSubscriber {
		return TokenResponse{}, ErrInvalidRole
	}

	roomName := RoomNameForDevice(request.DeviceID)
	identity := request.Identity
	if identity == "" {
		identity = DefaultParticipantIdentity(role, request.DeviceID)
	}

	validFor := request.ValidFor
	if validFor <= 0 {
		validFor = issuer.config.TokenTTL
	}

	expiresAt := time.Now().UTC().Add(validFor)

	canPublish := role == RolePublisher
	canSubscribe := true

	accessToken := auth.NewAccessToken(issuer.config.APIKey, issuer.config.APISecret)
	accessToken.SetIdentity(identity).
		SetValidFor(validFor).
		SetVideoGrant(&auth.VideoGrant{
			RoomJoin:     true,
			Room:         roomName,
			CanPublish:   &canPublish,
			CanSubscribe: &canSubscribe,
		})

	token, err := accessToken.ToJWT()
	if err != nil {
		return TokenResponse{}, fmt.Errorf("issue livekit token: %w", err)
	}

	return TokenResponse{
		Token:     token,
		URL:       issuer.config.URL,
		Room:      roomName,
		Identity:  identity,
		ExpiresAt: expiresAt,
	}, nil
}
