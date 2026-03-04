package auth

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const keyringService = "meta-cli"

type Store interface {
	GetTokens(account string) (*Tokens, error)
	SaveTokens(account string, t *Tokens) error
	GetSecret(account string) (string, error)
	SaveSecret(account string, secret string) error
}

type KeyringStore struct{}

func NewKeyringStore() *KeyringStore {
	return &KeyringStore{}
}

func (s *KeyringStore) GetTokens(account string) (*Tokens, error) {
	data, err := keyring.Get(keyringService, "tokens:"+account)
	if err != nil {
		return nil, fmt.Errorf("get tokens from keyring: %w", err)
	}
	var t Tokens
	if err := json.Unmarshal([]byte(data), &t); err != nil {
		return nil, fmt.Errorf("parse tokens: %w", err)
	}
	return &t, nil
}

func (s *KeyringStore) SaveTokens(account string, t *Tokens) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return keyring.Set(keyringService, "tokens:"+account, string(data))
}

func (s *KeyringStore) GetSecret(account string) (string, error) {
	return keyring.Get(keyringService, "secret:"+account)
}

func (s *KeyringStore) SaveSecret(account string, secret string) error {
	return keyring.Set(keyringService, "secret:"+account, secret)
}

type MemStore struct {
	tokens  map[string]*Tokens
	secrets map[string]string
}

func NewMemStore() *MemStore {
	return &MemStore{
		tokens:  make(map[string]*Tokens),
		secrets: make(map[string]string),
	}
}

func (s *MemStore) GetTokens(account string) (*Tokens, error) {
	t, ok := s.tokens[account]
	if !ok {
		return nil, fmt.Errorf("no tokens for account %q", account)
	}
	return t, nil
}

func (s *MemStore) SaveTokens(account string, t *Tokens) error {
	s.tokens[account] = t
	return nil
}

func (s *MemStore) GetSecret(account string) (string, error) {
	v, ok := s.secrets[account]
	if !ok {
		return "", fmt.Errorf("no secret for account %q", account)
	}
	return v, nil
}

func (s *MemStore) SaveSecret(account string, secret string) error {
	s.secrets[account] = secret
	return nil
}
