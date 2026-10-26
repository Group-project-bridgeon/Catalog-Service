package main

import (
	"strings"
	"testing"
)

func TestSelectionRejectsAmbiguousOrUnboundedInput(t *testing.T) {
	good := `{"unit_id":"11111111-1111-4111-8111-111111111111","variant_ids":["22222222-2222-4222-8222-222222222222"],"confirmation":"one base unit is one physical piece for every variant using this unit"}`
	if _, e := decodeSelection(strings.NewReader(good)); e != nil {
		t.Fatal("valid selection")
	}
	for _, bad := range []string{good + ` {}`, strings.Replace(good, `"unit_id":`, `"unit_id":null,"unit_id":`, 1), strings.Replace(good, `"confirmation":`, `"unexpected":true,"confirmation":`, 1), strings.Replace(good, "unit_id", "Unit_ID", 1), good + strings.Repeat(" ", 64<<10)} {
		if _, e := decodeSelection(strings.NewReader(bad)); e == nil {
			t.Fatal("ambiguous/unbounded selection accepted")
		}
	}
}
