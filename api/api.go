package api

import (
	"errors"
	"time"
)

type ErrorResponse struct {
	Op string
	Err string
}

func (e *ErrorResponse) Error() string {
	return e.Op + ": " + e.Err
}


type SetParams struct {
	Key string
	Value interface{}
	Ttl time.Duration
}

func ValidateSetParams(p *SetParams) error {
	if p.Key == "" {
		return errors.New("key argument must be specified")
	}
	if p.Value == nil {
		return errors.New("value argument must be specified")
	}
	if p.Ttl < 0 {
		return errors.New("ttl argument must be nonegative")
	}
	return nil
}


type GetParams struct {
	Key string
}

func ValidateGetParams(p *GetParams) error {
	if p.Key == "" {
		return errors.New("key argument must be specified")
	}
	return nil
}


type DelParams struct {
	Keys []string
}

func ValidateDelParams(p *DelParams) error {
	if p.Keys == nil || len(p.Keys) == 0 {
		return errors.New("at least one key must be in keys argument")
	}
	return nil
}


type KeysParams struct {
	Pattern string
}

func ValidateKeysParams(p *KeysParams) error {
	if p.Pattern == "" {
		return errors.New("pattern argument must be specified")
	}
	return nil
}