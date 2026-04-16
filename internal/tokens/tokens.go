package tokens

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Kind represents the type of honeytoken.
type Kind string

const (
	KindAPIKey      Kind = "api_key"
	KindCredentials Kind = "credentials"
	KindDatabase    Kind = "database_entry"
	KindDocument    Kind = "document"
	KindAWSCreds    Kind = "aws_credentials"
)

// Token represents a honeytoken — a fake credential or asset planted to detect unauthorized access.
type Token struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Kind            Kind           `json:"kind"`
	Value           string         `json:"value"`
	Description     string         `json:"description"`
	FirstAccessedAt *time.Time     `json:"first_accessed_at,omitempty"`
	LastAccessedAt  *time.Time     `json:"last_accessed_at,omitempty"`
	Active          bool           `json:"active"`
	CreatedAt       time.Time      `json:"created_at"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// Generator creates realistic-looking honeytokens.
type Generator struct {
	prefixes map[Kind]string
}

// NewGenerator creates a honeytoken generator with realistic prefixes.
func NewGenerator() *Generator {
	return &Generator{
		prefixes: map[Kind]string{
			KindAPIKey:      "sk-proj-",
			KindCredentials: "usr_",
			KindDatabase:    "db_",
			KindDocument:    "doc_",
			KindAWSCreds:    "AKIA",
		},
	}
}

// Generate creates a new honeytoken of the given kind.
func (g *Generator) Generate(kind Kind, name string, description string) Token {
	value := g.generateValue(kind)
	return Token{
		ID:          generateTokenID(),
		Name:        name,
		Kind:        kind,
		Value:       value,
		Description: description,
		Active:      true,
		CreatedAt:   time.Now().UTC(),
		Metadata:    map[string]any{},
	}
}

// GenerateBatch creates multiple honeytokens of the same kind.
func (g *Generator) GenerateBatch(kind Kind, prefix string, count int, description string) []Token {
	tokens := make([]Token, 0, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-%03d", prefix, i+1)
		tokens = append(tokens, g.Generate(kind, name, description))
	}
	return tokens
}

func (g *Generator) generateValue(kind Kind) string {
	randomBytes := make([]byte, 24)
	_, _ = rand.Read(randomBytes)
	randomPart := hex.EncodeToString(randomBytes)

	prefix, ok := g.prefixes[kind]
	if !ok {
		prefix = "htk_"
	}

	switch kind {
	case KindAPIKey:
		// Looks like an OpenAI API key: sk-proj-<48 hex chars>
		return prefix + randomPart[:48]
	case KindCredentials:
		// Looks like a service username: usr_<hex>
		return prefix + randomPart[:16]
	case KindDatabase:
		// Looks like a database connection string placeholder
		return fmt.Sprintf("postgres://admin:%s@db.internal.honeytrap:5432/secrets", randomPart[:24])
	case KindDocument:
		// Looks like a document URL
		return fmt.Sprintf("https://internal.honeytrap.local/docs/%s", randomPart[:16])
	case KindAWSCreds:
		// Looks like AWS access key ID (AKIA + 16 uppercase alphanumeric)
		chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		suffix := make([]byte, 16)
		for i := range suffix {
			suffix[i] = chars[randomBytes[i]%byte(len(chars))]
		}
		return prefix + string(suffix)
	default:
		return prefix + randomPart[:32]
	}
}

func generateTokenID() string {
	var raw [8]byte
	_, _ = rand.Read(raw[:])
	return "htk-" + hex.EncodeToString(raw[:])
}

// Store manages honeytokens with access tracking.
type Store struct {
	tokens map[string]Token
}

// NewStore creates a new token store.
func NewStore() *Store {
	return &Store{
		tokens: make(map[string]Token),
	}
}

// Add stores a token in the store.
func (s *Store) Add(token Token) {
	s.tokens[token.ID] = token
}

// Get retrieves a token by ID.
func (s *Store) Get(id string) (Token, bool) {
	t, ok := s.tokens[id]
	return t, ok
}

// GetByValue retrieves a token by its value (for access detection).
func (s *Store) GetByValue(value string) (Token, bool) {
	for _, t := range s.tokens {
		if t.Value == value {
			return t, true
		}
	}
	return Token{}, false
}

// List returns all tokens, optionally filtered by kind and active status.
func (s *Store) List(kind Kind, activeOnly bool) []Token {
	result := make([]Token, 0)
	for _, t := range s.tokens {
		if kind != "" && t.Kind != kind {
			continue
		}
		if activeOnly && !t.Active {
			continue
		}
		result = append(result, t)
	}
	return result
}

// RecordAccess marks a token as accessed, triggering an alert.
func (s *Store) RecordAccess(ctx context.Context, tokenID string) (Token, error) {
	t, ok := s.tokens[tokenID]
	if !ok {
		return Token{}, fmt.Errorf("token %s not found", tokenID)
	}
	now := time.Now().UTC()
	if t.FirstAccessedAt == nil {
		t.FirstAccessedAt = &now
	}
	t.LastAccessedAt = &now
	s.tokens[tokenID] = t
	return t, nil
}

// Deactivate marks a token as inactive.
func (s *Store) Deactivate(tokenID string) error {
	t, ok := s.tokens[tokenID]
	if !ok {
		return fmt.Errorf("token %s not found", tokenID)
	}
	t.Active = false
	s.tokens[tokenID] = t
	return nil
}