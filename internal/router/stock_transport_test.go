package router

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/security"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type transportResolver struct{ calls int }

func (r *transportResolver) ResolveBatch(ctx context.Context, items []domain.StockReferenceRequest) (domain.StockReferenceBatch, error) {
	r.calls++
	if ctx.Err() != nil {
		return domain.StockReferenceBatch{}, ctx.Err()
	}
	result := domain.StockReferenceBatch{CheckedAt: time.Now()}
	for _, i := range items {
		result.Items = append(result.Items, domain.StockReferenceItem{VariantID: i.VariantID, Status: "variant_not_found"})
	}
	return result, nil
}
func TestStockReferenceSecuredTransport(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal("key")
	}
	der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	path := filepath.Join(t.TempDir(), "public.pem")
	if os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), 0600) != nil {
		t.Fatal("public key")
	}
	machine, err := security.LoadMachine(path, "erp-auth")
	if err != nil {
		t.Fatal(err)
	}
	human, err := security.Load(path, "erp-auth", "erp-api")
	if err != nil {
		t.Fatal(err)
	}
	resolver := &transportResolver{}
	r := New(nil, zerolog.Nop(), 1<<20, time.Second, human, StockReferenceTransport{Verifier: machine, Resolver: resolver})
	registerBusinessRoutes(r) // Gin panics on wildcard conflicts.
	id := uuid.NewString()
	now := time.Now().Unix()
	claims := jwt.MapClaims{"iss": "erp-auth", "aud": []string{security.MachineAudience}, "sub": "machine:" + id, "client_id": id, "token_kind": "machine", "permissions": []string{security.MachinePermission}, "jti": uuid.NewString(), "iat": now, "nbf": now, "exp": now + 300}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	valid := `{"items":[{"variant_id":"` + uuid.NewString() + `"}]}`
	request := func(method, body, auth string, tlsEnabled bool) int {
		req := httptest.NewRequest(method, StockReferencePath, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if auth != "" {
			req.Header.Set("Authorization", "Bearer "+auth)
		}
		if tlsEnabled {
			req.TLS = &tls.ConnectionState{}
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if method == "POST" && w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("cache policy")
		}
		return w.Code
	}
	if status := request("POST", valid, token, true); status != 200 || resolver.calls != 1 {
		t.Fatal("valid machine request failed")
	}
	if request("POST", valid, "", true) != 401 {
		t.Fatal("anonymous accepted")
	}
	if request("POST", valid, token, false) != 400 {
		t.Fatal("plaintext accepted")
	}
	for _, body := range []string{valid + ` {}`, `{"items":[],"items":[]}`, `{"items":[],"Items":[]}`, `{"items":null}`, `{"items":[{"variant_id":null}]}`, `{"items":[{"variant_id":"bad"}]}`, `{"items":[{"variant_id":"` + uuid.NewString() + `","unexpected":1}]}`, strings.Repeat("x", (64<<10)+1)} {
		if request("POST", body, token, true) != 400 {
			t.Fatal("invalid request accepted")
		}
	}
	claims["permissions"] = []string{}
	denied, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if request("POST", valid, denied, true) != 403 {
		t.Fatal("missing authority accepted")
	}
	humanClaims := jwt.MapClaims{"iss": "erp-auth", "aud": []string{"erp-api"}, "sub": id, "user_id": id, "email": "synthetic@example.invalid", "token_version": 0, "permissions": []string{security.MachinePermission}, "iat": now, "nbf": now, "exp": now + 300}
	humanToken, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, humanClaims).SignedString(key)
	if request("POST", valid, humanToken, true) != 401 {
		t.Fatal("human token entered machine route")
	}
	req := httptest.NewRequest("GET", "/api/v1/catalog/units", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatal("machine entered human route")
	}
}
