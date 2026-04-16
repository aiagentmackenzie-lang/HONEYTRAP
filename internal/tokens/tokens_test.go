package tokens

import "testing"

func TestGenerateAPIKey(t *testing.T) {
	gen := NewGenerator()
	token := gen.Generate(KindAPIKey, "test-api-key", "Test API key for detection")

	if token.ID == "" {
		t.Error("expected token ID to be set")
	}
	if token.Kind != KindAPIKey {
		t.Errorf("expected kind %s, got %s", KindAPIKey, token.Kind)
	}
	if len(token.Value) < 20 {
		t.Errorf("expected value length >= 20, got %d", len(token.Value))
	}
	if token.Value[:8] != "sk-proj-" {
		t.Errorf("expected API key prefix 'sk-proj-', got %s", token.Value[:8])
	}
	if !token.Active {
		t.Error("expected token to be active")
	}
}

func TestGenerateCredentials(t *testing.T) {
	gen := NewGenerator()
	token := gen.Generate(KindCredentials, "test-user", "Test credential")

	if token.Value[:4] != "usr_" {
		t.Errorf("expected credential prefix 'usr_', got %s", token.Value[:4])
	}
}

func TestGenerateAWSCreds(t *testing.T) {
	gen := NewGenerator()
	token := gen.Generate(KindAWSCreds, "aws-key", "AWS access key")

	if token.Value[:4] != "AKIA" {
		t.Errorf("expected AWS prefix 'AKIA', got %s", token.Value[:4])
	}
	if len(token.Value) != 20 {
		t.Errorf("expected AWS key length 20, got %d", len(token.Value))
	}
}

func TestGenerateBatch(t *testing.T) {
	gen := NewGenerator()
	tokens := gen.GenerateBatch(KindAPIKey, "batch-test", 5, "Batch test tokens")

	if len(tokens) != 5 {
		t.Errorf("expected 5 tokens, got %d", len(tokens))
	}
	for i, token := range tokens {
		if token.Kind != KindAPIKey {
			t.Errorf("token %d: expected kind %s, got %s", i, KindAPIKey, token.Kind)
		}
	}
}

func TestStoreAddAndGet(t *testing.T) {
	store := NewStore()
	gen := NewGenerator()
	token := gen.Generate(KindAPIKey, "store-test", "Store test")

	store.Add(token)

	retrieved, ok := store.Get(token.ID)
	if !ok {
		t.Error("expected to find token by ID")
	}
	if retrieved.Value != token.Value {
		t.Errorf("expected value %s, got %s", token.Value, retrieved.Value)
	}
}

func TestStoreGetByValue(t *testing.T) {
	store := NewStore()
	gen := NewGenerator()
	token := gen.Generate(KindAPIKey, "value-lookup", "Value lookup test")

	store.Add(token)

	retrieved, ok := store.GetByValue(token.Value)
	if !ok {
		t.Error("expected to find token by value")
	}
	if retrieved.ID != token.ID {
		t.Errorf("expected ID %s, got %s", token.ID, retrieved.ID)
	}
}

func TestStoreRecordAccess(t *testing.T) {
	store := NewStore()
	gen := NewGenerator()
	token := gen.Generate(KindAPIKey, "access-test", "Access test")

	store.Add(token)

	updated, err := store.RecordAccess(nil, token.ID)
	if err != nil {
		t.Fatalf("RecordAccess failed: %v", err)
	}
	if updated.FirstAccessedAt == nil {
		t.Error("expected FirstAccessedAt to be set")
	}
	if updated.LastAccessedAt == nil {
		t.Error("expected LastAccessedAt to be set")
	}
}

func TestStoreDeactivate(t *testing.T) {
	store := NewStore()
	gen := NewGenerator()
	token := gen.Generate(KindAPIKey, "deact-test", "Deactivation test")

	store.Add(token)

	err := store.Deactivate(token.ID)
	if err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}

	retrieved, _ := store.Get(token.ID)
	if retrieved.Active {
		t.Error("expected token to be deactivated")
	}
}

func TestStoreListFiltered(t *testing.T) {
	store := NewStore()
	gen := NewGenerator()

	apiToken := gen.Generate(KindAPIKey, "api-1", "API token")
	credToken := gen.Generate(KindCredentials, "cred-1", "Credential")

	store.Add(apiToken)
	store.Add(credToken)

	apiOnly := store.List(KindAPIKey, false)
	if len(apiOnly) != 1 {
		t.Errorf("expected 1 API token, got %d", len(apiOnly))
	}

	all := store.List("", false)
	if len(all) != 2 {
		t.Errorf("expected 2 total tokens, got %d", len(all))
	}
}