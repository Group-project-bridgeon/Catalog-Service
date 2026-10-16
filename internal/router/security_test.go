package router

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/handler"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/security"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func routeTokens(t *testing.T) (*security.Verifier, func([]string) string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "public.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)}), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := security.Load(path, "erp-auth", "erp-api")
	if err != nil {
		t.Fatal(err)
	}
	return v, func(perms []string) string {
		now := time.Now().Unix()
		id := uuid.NewString()
		if perms == nil {
			perms = []string{}
		}
		raw, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"user_id": id, "sub": id, "token_version": 0, "permissions": perms, "iss": "erp-auth", "aud": "erp-api", "iat": now, "nbf": now, "exp": now + 900}).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
}

func registerBusinessRoutes(r *gin.Engine) {
	RegisterCategoryRoutes(r, (*handler.CategoryHandler)(nil))
	RegisterBrandRoutes(r, (*handler.BrandHandler)(nil))
	RegisterUnitRoutes(r, (*handler.UnitHandler)(nil))
	RegisterProductRoutes(r, (*handler.ProductHandler)(nil))
	RegisterProductVariantRoutes(r, (*handler.ProductVariantHandler)(nil))
	RegisterVariantValueRoutes(r, (*handler.VariantValueHandler)(nil))
	RegisterAttributeRoutes(r, (*handler.AttributeHandler)(nil))
	RegisterAttributeValueRoutes(r, (*handler.AttributeValueHandler)(nil))
	RegisterProductImageRoutes(r, (*handler.ProductImageHandler)(nil))
	RegisterPriceListRoutes(r, (*handler.PriceListHandler)(nil))
	RegisterPriceRoutes(r, (*handler.PriceHandler)(nil))
	RegisterProductDetailRoutes(r, (*handler.ProductDetailHandler)(nil))
}

func TestEveryBusinessRouteEnforcesMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	v, sign := routeTokens(t)
	r := New(fakeDatabase{}, zerolog.Nop(), 1<<20, time.Second, v)
	// Sentinel runs only after the real router's security middleware. Nil
	// handlers must never execute; separate tests exercise actual handlers.
	r.Use(func(c *gin.Context) { c.AbortWithStatus(204) })
	registerBusinessRoutes(r)
	policy := PermissionMatrix()
	count := 0
	for _, route := range r.Routes() {
		if route.Path == "/health" || route.Path == "/ready" {
			continue
		}
		count++
		required, ok := policy[route.Method+" "+route.Path]
		if !ok || len(required) == 0 {
			t.Fatalf("unmapped route %s %s", route.Method, route.Path)
		}
		path := strings.NewReplacer(":variant_id", uuid.NewString(), ":attribute_id", uuid.NewString(), ":id", uuid.NewString()).Replace(route.Path)
		cases := []struct {
			name, token string
			code        int
		}{{"missing", "", 401}, {"empty", sign(nil), 403}, {"unrelated", sign([]string{"admin:all", "catalog:*"}), 403}, {"allowed", sign(required), 204}}
		for i := range required {
			partial := append([]string{}, required[:i]...)
			partial = append(partial, required[i+1:]...)
			cases = append(cases, struct {
				name, token string
				code        int
			}{"missing " + required[i], sign(partial), 403})
		}
		for _, tc := range cases {
			t.Run(route.Method+" "+route.Path+"/"+tc.name, func(t *testing.T) {
				req := httptest.NewRequest(route.Method, path, strings.NewReader(`{}`))
				if tc.token != "" {
					req.Header.Set("Authorization", "Bearer "+tc.token)
				}
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != tc.code {
					t.Fatalf("got %d want %d", w.Code, tc.code)
				}
			})
		}
	}
	if count != len(policy) || count != 58 {
		t.Fatalf("registered=%d policy=%d expected=58", count, len(policy))
	}
	// A newly added route without an explicit policy stays closed.
	r.GET("/new-business-route", func(c *gin.Context) { t.Error("unmapped handler executed") })
	req := httptest.NewRequest("GET", "/new-business-route", nil)
	req.Header.Set("Authorization", "Bearer "+sign([]string{"catalog.products:read"}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("unmapped route returned %d", w.Code)
	}
}

// Memory fixture verifies real HTTP handlers cannot mutate on 401 or 403.
type brandFixture struct {
	writes int
	brand  domain.Brand
}

func (f *brandFixture) Create(_ context.Context, in service.CreateBrandInput) (domain.Brand, error) {
	f.writes++
	f.brand = domain.Brand{ID: uuid.New(), Name: in.Name, Code: in.Code, IsActive: true}
	return f.brand, nil
}
func (f *brandFixture) GetByID(context.Context, uuid.UUID) (domain.Brand, error) { return f.brand, nil }
func (f *brandFixture) List(context.Context) ([]domain.Brand, error) {
	return []domain.Brand{f.brand}, nil
}
func (f *brandFixture) Update(_ context.Context, _ uuid.UUID, in service.UpdateBrandInput) (domain.Brand, error) {
	f.writes++
	f.brand.Name = in.Name
	f.brand.Code = in.Code
	return f.brand, nil
}
func (f *brandFixture) SetActive(_ context.Context, _ uuid.UUID, active bool) (domain.Brand, error) {
	f.writes++
	f.brand.IsActive = active
	return f.brand, nil
}

func TestAuthorizedHandlersAndDeniedWrites(t *testing.T) {
	v, sign := routeTokens(t)
	f := &brandFixture{}
	r := New(fakeDatabase{}, zerolog.Nop(), 1<<20, time.Second, v)
	RegisterBrandRoutes(r, handler.NewBrandHandler(f, zerolog.Nop()))
	request := func(method, path, body, token string, expected int) {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != expected {
			t.Fatalf("%s %s got %d want %d", method, path, w.Code, expected)
		}
	}
	base := "/api/v1/catalog/brands"
	body := `{"name":"Fixture","code":"FIXTURE"}`
	request("POST", base, body, "", 401)
	request("POST", base, body, sign([]string{"catalog.brands:read"}), 403)
	if f.writes != 0 {
		t.Fatal("denied create mutated data")
	}
	request("POST", base, body, sign([]string{"catalog.brands:create"}), 201)
	path := base + "/" + f.brand.ID.String()
	request("GET", path, "", sign([]string{"catalog.brands:read"}), 200)
	request("GET", base, "", sign([]string{"catalog.brands:read"}), 200)
	request("PUT", path, body, "", 401)
	request("PUT", path, body, sign(nil), 403)
	request("PATCH", path+"/status", `{"is_active":false}`, sign(nil), 403)
	if f.writes != 1 || !f.brand.IsActive {
		t.Fatal("denied update mutated data")
	}
	request("PUT", path, body, sign([]string{"catalog.brands:update"}), 200)
	request("PATCH", path+"/status", `{"is_active":false}`, sign([]string{"catalog.brands:update"}), 200)
	if f.writes != 3 || f.brand.IsActive {
		t.Fatal("authorized updates did not execute")
	}
	request("GET", "/health", "", "", 200)
	request("GET", "/ready", "", "", 200)
}
