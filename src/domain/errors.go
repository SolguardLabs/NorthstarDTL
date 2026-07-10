package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrInvalidIntent      ErrorCode = "invalid_intent"
	ErrRouteUnavailable   ErrorCode = "route_unavailable"
	ErrRouteNotFound      ErrorCode = "route_not_found"
	ErrInsufficientFunds  ErrorCode = "insufficient_funds"
	ErrInsufficientRoute  ErrorCode = "insufficient_route_liquidity"
	ErrExposureLimit      ErrorCode = "exposure_limit"
	ErrTicketNotFound     ErrorCode = "ticket_not_found"
	ErrTicketState        ErrorCode = "ticket_state"
	ErrDuplicateID        ErrorCode = "duplicate_id"
	ErrUnknownAction      ErrorCode = "unknown_action"
	ErrInvalidBootstrap   ErrorCode = "invalid_bootstrap"
	ErrSettlementRejected ErrorCode = "settlement_rejected"
)

type DomainError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e DomainError) Error() string {
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, format string, args ...any) DomainError {
	return DomainError{Code: code, Message: fmt.Sprintf(format, args...)}
}

func Invalid(format string, args ...any) DomainError {
	return NewError(ErrInvalidIntent, format, args...)
}

func NotFound(format string, args ...any) DomainError {
	return NewError(ErrRouteNotFound, format, args...)
}

func Duplicate(format string, args ...any) DomainError {
	return NewError(ErrDuplicateID, format, args...)
}

func InsufficientFunds(format string, args ...any) DomainError {
	return NewError(ErrInsufficientFunds, format, args...)
}

func InsufficientRoute(format string, args ...any) DomainError {
	return NewError(ErrInsufficientRoute, format, args...)
}

func ExposureLimit(format string, args ...any) DomainError {
	return NewError(ErrExposureLimit, format, args...)
}

func TicketState(format string, args ...any) DomainError {
	return NewError(ErrTicketState, format, args...)
}

func AsDomainError(err error) (DomainError, bool) {
	if err == nil {
		return DomainError{}, false
	}
	var direct DomainError
	if errors.As(err, &direct) {
		return direct, true
	}
	return DomainError{}, false
}
