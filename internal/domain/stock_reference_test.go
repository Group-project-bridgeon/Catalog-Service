package domain

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"math"
	"strings"
	"testing"
	"time"
)

func TestStockCountAndPackValidation(t *testing.T) {
	for _, s := range []string{"1", "9007199254740993", "9223372036854775807"} {
		n, e := ParseStockCount(s)
		if e != nil {
			t.Fatal(e)
		}
		b, e := json.Marshal(n)
		if e != nil || string(b) != `"`+s+`"` {
			t.Fatal("inexact JSON")
		}
		var out StockCount
		if json.Unmarshal(b, &out) != nil || n != out {
			t.Fatal("roundtrip")
		}
	}
	for _, s := range []string{"", "0", "-1", "+1", "01", "1.0", "1.5", "1e2", " 1", "9223372036854775808", "-9223372036854775808"} {
		if _, e := ParseStockCount(s); !errors.Is(e, ErrStockInvalid) {
			t.Fatal("invalid count accepted")
		}
	}
	for _, raw := range []string{`1`, `null`, `"1.0"`, `"0"`, `"9223372036854775808"`} {
		n := StockCount(7)
		if json.Unmarshal([]byte(raw), &n) == nil || n != 7 {
			t.Fatal("invalid serialization mutated count")
		}
	}
	for _, n := range []StockCount{0, -1, math.MinInt64} {
		if _, e := json.Marshal(n); e == nil {
			t.Fatal("nonpositive serialized")
		}
	}
	if code, label, e := NormalizePack(" box-6 ", " Box of six ", 6); e != nil || code != "BOX-6" || label != "Box of six" {
		t.Fatal("normalization")
	}
	for _, tc := range []struct {
		code, label string
		factor      StockCount
	}{{"BAD CODE", "Label", 1}, {"OK", "", 1}, {"OK", strings.Repeat("x", 201), 1}, {"OK", "bad\nlabel", 1}, {"OK", "Label", 0}} {
		if _, _, e := NormalizePack(tc.code, tc.label, tc.factor); e == nil {
			t.Fatal("invalid pack accepted")
		}
	}
}
func stockReferenceFixture() StockReference {
	return StockReference{ProductID: uuid.New(), VariantID: uuid.New(), BaseUnitID: uuid.New(), QuantityKind: "PIECE", Eligible: true, ProductActive: true, VariantActive: true, BaseUnitActive: true, ProductCode: "P", ProductName: "Product", SKU: "SKU", BaseUnitCode: "PC", BaseUnitName: "Piece", ProductVersion: 1, VariantVersion: 1, UnitVersion: 1, CheckedAt: time.Now().UTC()}
}
func TestStockFingerprintFields(t *testing.T) {
	r := stockReferenceFixture()
	r.Pack = &VariantPack{ID: uuid.New(), VariantID: r.VariantID, Code: "BOX", Revision: 1, CurrentRevision: 1, Version: 1, PiecesPerPack: 6, Label: "Six", Active: true, Current: true}
	want, e := r.Fingerprint()
	if e != nil || len(want) != 77 {
		t.Fatal("revision contract")
	}
	copy := r
	copy.CheckedAt = r.CheckedAt.Add(time.Hour)
	copy.Revision = "ignored"
	if got, _ := copy.Fingerprint(); got != want {
		t.Fatal("volatile revision input")
	}
	for _, change := range []func(*StockReference){func(r *StockReference) { r.ProductID = uuid.New() }, func(r *StockReference) { r.VariantID = uuid.New() }, func(r *StockReference) { r.BaseUnitID = uuid.New() }, func(r *StockReference) { r.Eligible = false }, func(r *StockReference) { r.QuantityKind = "UNCLASSIFIED" }, func(r *StockReference) { r.ProductActive = false }, func(r *StockReference) { r.VariantActive = false }, func(r *StockReference) { r.BaseUnitActive = false }, func(r *StockReference) { r.ProductCode += "x" }, func(r *StockReference) { r.ProductName += "x" }, func(r *StockReference) { r.SKU += "x" }, func(r *StockReference) { r.BaseUnitCode += "x" }, func(r *StockReference) { r.BaseUnitName += "x" }, func(r *StockReference) { r.ProductVersion++ }, func(r *StockReference) { r.VariantVersion++ }, func(r *StockReference) { r.UnitVersion++ }, func(r *StockReference) { r.Pack.ID = uuid.New() }, func(r *StockReference) { r.Pack.VariantID = uuid.New() }, func(r *StockReference) { r.Pack.Code += "X" }, func(r *StockReference) { r.Pack.Revision++ }, func(r *StockReference) { r.Pack.CurrentRevision++ }, func(r *StockReference) { r.Pack.Version++ }, func(r *StockReference) { r.Pack.PiecesPerPack++ }, func(r *StockReference) { r.Pack.Label += "x" }, func(r *StockReference) { r.Pack.Active = false }, func(r *StockReference) { r.Pack.Current = false }, func(r *StockReference) { r.Pack = nil }} {
		c := r
		p := *r.Pack
		c.Pack = &p
		change(&c)
		if got, e := c.Fingerprint(); e != nil || got == want {
			t.Fatal("meaningful field not versioned")
		}
	}
}
func TestStockRequestBounds(t *testing.T) {
	req := StockReferenceRequest{VariantID: uuid.New()}
	if ValidateStockRequests([]StockReferenceRequest{req}) != nil {
		t.Fatal("valid request")
	}
	for _, items := range [][]StockReferenceRequest{nil, {req, req}, {{}}, {{VariantID: req.VariantID, PackRevision: "1"}}, {{VariantID: req.VariantID, PackID: uuid.New(), PackRevision: "01"}}, {{VariantID: req.VariantID, ExpectedRevision: "timestamp"}}} {
		if ValidateStockRequests(items) == nil {
			t.Fatal("invalid requests accepted")
		}
	}
	items := make([]StockReferenceRequest, 100)
	for i := range items {
		items[i].VariantID = uuid.New()
	}
	if ValidateStockRequests(items) != nil {
		t.Fatal("100 rejected")
	}
	items = append(items, req)
	if ValidateStockRequests(items) == nil {
		t.Fatal("101 accepted")
	}
}

func TestStockFingerprintV1Golden(t *testing.T) {
	r := StockReference{ProductID: uuid.MustParse("01935f00-0000-7000-8000-000000000001"), VariantID: uuid.MustParse("01935f00-0000-7000-8000-000000000002"), BaseUnitID: uuid.MustParse("01935f00-0000-7000-8000-000000000003"), QuantityKind: "PIECE", Eligible: true, ProductActive: true, VariantActive: true, BaseUnitActive: true, ProductCode: "P", ProductName: "Product", SKU: "SKU", BaseUnitCode: "PC", BaseUnitName: "Piece", ProductVersion: 1, VariantVersion: 1, UnitVersion: 1}
	// Independently hashed with .NET SHA256 over the literal ordered canonical
	// JSON fields documented in stock-reference-spec.md. Do not regenerate on drift.
	const want = "cat-stock-v1:dee7899214f51b1d0441c493ad3592db2cbb78740b081bbfcde3710fd73dffcd"
	got, err := r.Fingerprint()
	if err != nil || got != want {
		t.Fatal("canonical revision v1 drift")
	}
	wire, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if json.Unmarshal(wire, &decoded) != nil {
		t.Fatal("invalid reference JSON")
	}
	for _, field := range []string{"product_version", "variant_version", "unit_version"} {
		if string(decoded[field]) != `"1"` {
			t.Fatal("int64 metadata serialized as JSON number")
		}
	}
}
