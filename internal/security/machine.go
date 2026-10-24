package security

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const MachineAudience = "catalog-stock-references"
const MachinePermission = "catalog.stock-references:read"
const MachineMaxLifetime = 5 * time.Minute
const MachineClockTolerance = 30 * time.Second

type MachineClaims struct {
	ClientID    string   `json:"client_id"`
	Kind        string   `json:"token_kind"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// Reject duplicate keys and require exact integer NumericDates before jwt/v5's
// more permissive decoder. Unknown claims are rejected for this versioned contract.
func (c *MachineClaims) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	start, err := dec.Token()
	if err != nil || start != json.Delim('{') {
		return ErrInvalidToken
	}
	fields := map[string]json.RawMessage{}
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return ErrInvalidToken
		}
		name, ok := key.(string)
		if !ok {
			return ErrInvalidToken
		}
		if _, exists := fields[name]; exists {
			return ErrInvalidToken
		}
		var value json.RawMessage
		if dec.Decode(&value) != nil {
			return ErrInvalidToken
		}
		fields[name] = value
	}
	if _, err = dec.Token(); err != nil {
		return ErrInvalidToken
	}
	if _, err = dec.Token(); err != io.EOF {
		return ErrInvalidToken
	}
	if len(fields) != 10 {
		return ErrInvalidToken
	}
	for _, name := range []string{"iss", "sub", "client_id", "token_kind", "jti"} {
		var value string
		if json.Unmarshal(fields[name], &value) != nil || value == "" {
			return ErrInvalidToken
		}
	}
	for _, name := range []string{"iat", "nbf", "exp"} {
		var value int64
		if bytes.Equal(fields[name], []byte("null")) || json.Unmarshal(fields[name], &value) != nil || value <= 0 {
			return ErrInvalidToken
		}
	}
	var aud, permissions []string
	if json.Unmarshal(fields["aud"], &aud) != nil || len(aud) != 1 || aud[0] != MachineAudience {
		return ErrInvalidToken
	}
	if json.Unmarshal(fields["permissions"], &permissions) != nil || permissions == nil {
		return ErrInvalidToken
	}
	type plain MachineClaims
	return json.Unmarshal(data, (*plain)(c))
}
func (c MachineClaims) Validate() error {
	id, err := uuid.Parse(c.ClientID)
	if err != nil || id == uuid.Nil || id.String() != c.ClientID || c.Subject != "machine:"+c.ClientID || c.Kind != "machine" {
		return ErrInvalidToken
	}
	jti, err := uuid.Parse(c.ID)
	if err != nil || jti == uuid.Nil || jti.String() != c.ID {
		return ErrInvalidToken
	}
	if c.IssuedAt == nil || c.NotBefore == nil || c.ExpiresAt == nil {
		return ErrInvalidToken
	}
	if !c.NotBefore.Equal(c.IssuedAt.Time) || !c.ExpiresAt.After(c.IssuedAt.Time) || c.ExpiresAt.Sub(c.IssuedAt.Time) > MachineMaxLifetime {
		return ErrInvalidToken
	}
	if len(c.Permissions) > 1 || (len(c.Permissions) == 1 && c.Permissions[0] != MachinePermission) {
		return ErrInvalidToken
	}
	return nil
}

type MachineVerifier struct {
	key    *rsa.PublicKey
	issuer string
}

func LoadMachine(path, issuer string) (*MachineVerifier, error) {
	v, err := Load(path, issuer, MachineAudience)
	if err != nil {
		return nil, err
	}
	return &MachineVerifier{v.key, v.issuer}, nil
}
func (v *MachineVerifier) Verify(raw string) (*MachineClaims, error) {
	if v == nil || v.key == nil || len(raw) > 8192 {
		return nil, ErrInvalidToken
	}
	token, err := jwt.ParseWithClaims(raw, &MachineClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, ErrInvalidToken
		}
		return v.key, nil
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(v.issuer), jwt.WithAudience(MachineAudience), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(MachineClockTolerance))
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	c, ok := token.Claims.(*MachineClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return c, nil
}

// ProtectMachine is deliberately not registered on any production route in 4b.
func ProtectMachine(v *MachineVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := c.Request.Header.Values("Authorization")
		if len(headers) != 1 {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid bearer token"})
			return
		}
		parts := strings.Fields(headers[0])
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid bearer token"})
			return
		}
		claims, err := v.Verify(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid bearer token"})
			return
		}
		if len(claims.Permissions) != 1 || claims.Permissions[0] != MachinePermission {
			c.AbortWithStatusJSON(403, gin.H{"error": "permission denied"})
			return
		}
		c.Set("catalog.verifiedMachine", claims)
		c.Next()
	}
}
