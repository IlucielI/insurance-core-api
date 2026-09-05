package constants

import (
	"errors"
	"testing"
)

func TestIsQuoteValidationError(t *testing.T) {
	if !IsQuoteValidationError(QuoteSumAssuredOutOfRangeError) {
		t.Fatal("IsQuoteValidationError(sum assured) = false, want true")
	}
	if !IsQuoteValidationError(QuotePaymentTermOutOfRangeError) {
		t.Fatal("IsQuoteValidationError(payment term) = false, want true")
	}
	if IsQuoteValidationError(QuotePricingRulesInvalidError) {
		t.Fatal("IsQuoteValidationError(pricing rules) = true, want false")
	}
	if IsQuoteValidationError(errors.New("other")) {
		t.Fatal("IsQuoteValidationError(other) = true, want false")
	}
}

func TestProductQuoteNotes(t *testing.T) {
	notes := ProductQuoteNotes()
	if len(notes) != 2 {
		t.Fatalf("ProductQuoteNotes() length = %d, want 2", len(notes))
	}
	notes[0] = "changed"
	if ProductQuoteNotes()[0] == "changed" {
		t.Fatal("ProductQuoteNotes() returned shared mutable slice")
	}
}
