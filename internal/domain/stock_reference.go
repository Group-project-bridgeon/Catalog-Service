package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrStockInvalid        = errors.New("invalid Catalog stock input")
	ErrStockNotFound       = errors.New("Catalog stock reference not found")
	ErrStockConflict       = errors.New("Catalog stock version or code conflict")
	ErrStockImmutable      = errors.New("Catalog stock meaning is immutable")
	ErrStockIneligible     = errors.New("Catalog stock classification invalid")
	ErrStockBusy           = errors.New("Catalog stock operation busy")
	ErrStockInternal       = errors.New("Catalog stock operation failed")
	ErrStockOutcomeUnknown = errors.New("Catalog management commit outcome unknown; reconcile before retry")
)

// StockCount is a positive exact int64, serialized as a canonical decimal string.
// It represents factors and revisions, never price or fractional quantities.
type StockCount int64

func ParseStockCount(s string) (StockCount, error) {
	if s == "" || s[0] < '1' || s[0] > '9' {
		return 0, ErrStockInvalid
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrStockInvalid
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrStockInvalid
	}
	return StockCount(n), nil
}
func (n StockCount) MarshalJSON() ([]byte, error) {
	if n <= 0 {
		return nil, ErrStockInvalid
	}
	return json.Marshal(strconv.FormatInt(int64(n), 10))
}
func (n *StockCount) UnmarshalJSON(b []byte) error {
	if n == nil {
		return ErrStockInvalid
	}
	var s string
	if json.Unmarshal(b, &s) != nil {
		return ErrStockInvalid
	}
	v, err := ParseStockCount(s)
	if err == nil {
		*n = v
	}
	return err
}
func StockText(s string, max int) bool {
	if !utf8.ValidString(s) || strings.TrimSpace(s) == "" || utf8.RuneCountInString(s) > max {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

var packCode = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{0,49}$`)

func NormalizePack(code, label string, factor StockCount) (string, string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	label = strings.TrimSpace(label)
	if !packCode.MatchString(code) || !StockText(label, 200) || factor <= 0 {
		return "", "", ErrStockInvalid
	}
	return code, label, nil
}

type VariantPack struct {
	ID              uuid.UUID  `json:"id"`
	VariantID       uuid.UUID  `json:"variant_id"`
	Code            string     `json:"code"`
	Revision        StockCount `json:"revision"`
	CurrentRevision StockCount `json:"current_revision"`
	Version         StockCount `json:"version"`
	PiecesPerPack   StockCount `json:"pieces_per_pack"`
	Label           string     `json:"label"`
	Active          bool       `json:"is_active"`
	Current         bool       `json:"is_current"`
}
type StockReferenceRequest struct {
	VariantID        uuid.UUID `json:"variant_id"`
	PackID           uuid.UUID `json:"pack_id,omitempty"`
	PackRevision     string    `json:"pack_revision,omitempty"`
	ExpectedRevision string    `json:"expected_revision,omitempty"`
}

func ValidateStockRequests(items []StockReferenceRequest) error {
	if len(items) < 1 || len(items) > 100 {
		return ErrStockInvalid
	}
	seen := map[uuid.UUID]bool{}
	for _, i := range items {
		if i.VariantID == uuid.Nil || seen[i.VariantID] {
			return ErrStockInvalid
		}
		seen[i.VariantID] = true
		if i.PackID == uuid.Nil {
			if i.PackRevision != "" {
				return ErrStockInvalid
			}
		} else if _, err := ParseStockCount(i.PackRevision); err != nil {
			return err
		}
		if i.ExpectedRevision != "" {
			if len(i.ExpectedRevision) != 77 || !strings.HasPrefix(i.ExpectedRevision, "cat-stock-v1:") {
				return ErrStockInvalid
			}
			if _, err := hex.DecodeString(i.ExpectedRevision[13:]); err != nil {
				return ErrStockInvalid
			}
		}
	}
	return nil
}

type StockReference struct {
	ProductID      uuid.UUID    `json:"product_id"`
	VariantID      uuid.UUID    `json:"variant_id"`
	BaseUnitID     uuid.UUID    `json:"base_unit_id"`
	QuantityKind   string       `json:"quantity_kind"`
	Eligible       bool         `json:"eligible"`
	ProductActive  bool         `json:"product_active"`
	VariantActive  bool         `json:"variant_active"`
	BaseUnitActive bool         `json:"base_unit_active"`
	ProductCode    string       `json:"product_code"`
	ProductName    string       `json:"product_name"`
	SKU            string       `json:"sku"`
	BaseUnitCode   string       `json:"base_unit_code"`
	BaseUnitName   string       `json:"base_unit_name"`
	ProductVersion StockCount   `json:"product_version"`
	VariantVersion StockCount   `json:"variant_version"`
	UnitVersion    StockCount   `json:"unit_version"`
	Pack           *VariantPack `json:"pack,omitempty"`
	Revision       string       `json:"stock_reference_revision"`
	CheckedAt      time.Time    `json:"checked_at"`
}

// Explicit v1 canonical fields. No caller expectations, clocks, maps or floats.
func (r StockReference) Fingerprint() (string, error) {
	canonical := struct {
		Version                                            string
		ProductID, VariantID, BaseUnitID                   uuid.UUID
		Kind                                               string
		Eligible, ProductActive, VariantActive, UnitActive bool
		ProductCode, ProductName, SKU, UnitCode, UnitName  string
		ProductVersion, VariantVersion, UnitVersion        StockCount
		Pack                                               *VariantPack
	}{"catalog.stock-reference.v1", r.ProductID, r.VariantID, r.BaseUnitID, r.QuantityKind, r.Eligible, r.ProductActive, r.VariantActive, r.BaseUnitActive, r.ProductCode, r.ProductName, r.SKU, r.BaseUnitCode, r.BaseUnitName, r.ProductVersion, r.VariantVersion, r.UnitVersion, r.Pack}
	b, err := json.Marshal(canonical)
	if err != nil {
		return "", ErrStockInvalid
	}
	h := sha256.Sum256(b)
	return "cat-stock-v1:" + hex.EncodeToString(h[:]), nil
}
func (r StockReference) ValidMetadata() bool {
	return r.ProductID != uuid.Nil && r.VariantID != uuid.Nil && r.BaseUnitID != uuid.Nil && StockText(r.ProductCode, 50) && StockText(r.ProductName, 200) && StockText(r.SKU, 100) && StockText(r.BaseUnitCode, 30) && StockText(r.BaseUnitName, 100)
}

type StockReferenceItem struct {
	VariantID uuid.UUID       `json:"variant_id"`
	Status    string          `json:"status"`
	Reference *StockReference `json:"reference,omitempty"`
}
type StockReferenceBatch struct {
	CheckedAt time.Time            `json:"checked_at"`
	Items     []StockReferenceItem `json:"items"`
}
