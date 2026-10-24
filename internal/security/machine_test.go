package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func machineFixture() jwt.MapClaims {
	now := time.Now().Unix()
	id := uuid.NewString()
	return jwt.MapClaims{"iss": "erp-auth", "aud": []string{MachineAudience}, "sub": "machine:" + id, "client_id": id, "token_kind": "machine", "permissions": []string{MachinePermission}, "jti": uuid.NewString(), "iat": now, "nbf": now, "exp": now + 300}
}
func TestMachineClaimsAndSeparation(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal("key generation failed")
	}
	v := &MachineVerifier{&key.PublicKey, "erp-auth"}
	sign := func(c jwt.MapClaims) string {
		raw, err := jwt.NewWithClaims(jwt.SigningMethodRS256, c).SignedString(key)
		if err != nil {
			t.Fatal("sign failed")
		}
		return raw
	}
	if _, err = v.Verify(sign(machineFixture())); err != nil {
		t.Fatal("valid rejected")
	}
	cases := map[string]func(jwt.MapClaims){
		"issuer": func(c jwt.MapClaims) { c["iss"] = "other" }, "audience": func(c jwt.MapClaims) { c["aud"] = []string{"erp-api"} }, "extra audience": func(c jwt.MapClaims) { c["aud"] = []string{MachineAudience, "erp-api"} }, "kind": func(c jwt.MapClaims) { c["token_kind"] = "human" }, "subject": func(c jwt.MapClaims) { c["sub"] = c["client_id"] }, "zero client": func(c jwt.MapClaims) { c["client_id"] = uuid.Nil.String() }, "unknown": func(c jwt.MapClaims) { c["user_id"] = uuid.NewString() }, "bad permission": func(c jwt.MapClaims) { c["permissions"] = []string{"catalog.products:read"} }, "null permissions": func(c jwt.MapClaims) { c["permissions"] = nil }, "string date": func(c jwt.MapClaims) { c["iat"] = "1" }, "fraction": func(c jwt.MapClaims) { c["iat"] = 1.5 }, "overlong despite skew": func(c jwt.MapClaims) { c["exp"] = c["iat"].(int64) + 301 }, "expired": func(c jwt.MapClaims) { n := time.Now().Unix() - 400; c["iat"] = n; c["nbf"] = n; c["exp"] = n + 300 }, "future": func(c jwt.MapClaims) { n := time.Now().Unix() + 60; c["iat"] = n; c["nbf"] = n; c["exp"] = n + 300 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			c := machineFixture()
			change(c)
			if _, err := v.Verify(sign(c)); err == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
	for field := range machineFixture() {
		for _, null := range []bool{false, true} {
			c := machineFixture()
			if null {
				c[field] = nil
			} else {
				delete(c, field)
			}
			if _, err := v.Verify(sign(c)); err == nil {
				t.Fatalf("absent/null %s accepted", field)
			}
		}
	}
	c := machineFixture()
	n := time.Now().Unix() - 310
	c["iat"] = n
	c["nbf"] = n
	c["exp"] = n + 300
	if _, err := v.Verify(sign(c)); err != nil {
		t.Fatal("within expiry tolerance rejected")
	}
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	wrong, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, machineFixture()).SignedString(other)
	if _, err := v.Verify(wrong); err == nil {
		t.Fatal("wrong signature accepted")
	}
	hs, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, machineFixture()).SignedString([]byte("synthetic-invalid-algorithm"))
	if _, err := v.Verify(hs); err == nil {
		t.Fatal("HS accepted")
	}
	human := &Verifier{key: &key.PublicKey, issuer: "erp-auth", audience: "erp-api"}
	if _, err := human.Verify(sign(machineFixture())); err == nil {
		t.Fatal("human verifier accepted machine")
	}
	h := jwt.MapClaims{"iss": "erp-auth", "aud": []string{"erp-api"}, "sub": uuid.NewString(), "user_id": uuid.NewString(), "token_version": 0, "permissions": []string{MachinePermission}, "iat": time.Now().Unix(), "nbf": time.Now().Unix(), "exp": time.Now().Unix() + 300}
	if _, err := v.Verify(sign(h)); err == nil {
		t.Fatal("machine verifier accepted human")
	}
	c = machineFixture()
	c["permissions"] = []string{}
	r := gin.New()
	r.GET("/test", ProtectMachine(v), func(c *gin.Context) { c.Status(204) })
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+sign(c))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatal("missing permission not forbidden")
	}
	var parsed MachineClaims
	if json.Unmarshal([]byte(`{"iss":"a","iss":"b"}`), &parsed) == nil {
		t.Fatal("duplicate claim accepted")
	}
}

// The parent Auth integration test sends a real HTTP-issued token over stdin;
// no token or private key is saved to disk or included in test output.
func TestMachineIssuerContract(t *testing.T) {
	if os.Getenv("AUTH_MACHINE_CONTRACT_STDIN") != "1" {
		t.Skip("requires real Auth issuer fixture")
	}
	var input struct {
		Token     string
		PublicKey string
	}
	if json.NewDecoder(os.Stdin).Decode(&input) != nil {
		t.Fatal("fixture input invalid")
	}
	der, err := base64.StdEncoding.DecodeString(input.PublicKey)
	if err != nil {
		t.Fatal("fixture public key invalid")
	}
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		t.Fatal("fixture public key invalid")
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		t.Fatal("not RSA")
	}
	v := &MachineVerifier{rsaKey, "erp-auth"}
	claims, err := v.Verify(input.Token)
	if err != nil || len(claims.Permissions) != 1 || claims.Permissions[0] != MachinePermission {
		t.Fatal("actual issuer contract rejected")
	}
	human := &Verifier{key: rsaKey, issuer: "erp-auth", audience: "erp-api"}
	if _, err := human.Verify(input.Token); err == nil {
		t.Fatal("human boundary accepted issued machine token")
	}
	r := gin.New()
	r.GET("/test-only", ProtectMachine(v), func(c *gin.Context) { c.Status(204) })
	req := httptest.NewRequest("GET", "/test-only", nil)
	req.Header.Set("Authorization", "Bearer "+input.Token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatal("machine authorization failed")
	}
}
