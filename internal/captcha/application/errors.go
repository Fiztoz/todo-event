package application

import "errors"

var (
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrChallengeExpired     = errors.New("challenge expired")
	ErrChallengeAlreadyUsed = errors.New("challenge already verified")
	ErrIncorrectAnswer      = errors.New("incorrect answer")
)
