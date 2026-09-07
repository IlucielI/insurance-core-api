package email

import (
	"math"
	"strings"
	"testing"
)

func TestRendererAllTemplates(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	t.Run("ApplicationSubmitted", func(t *testing.T) {
		text, html, err := renderer.RenderApplicationSubmitted(ApplicationSubmittedData{
			FullName:            "Bayu Pratama",
			ProductName:         "Secure Life Plus",
			ApplicationID:       "APP-2026-8819",
			SumAssuredFormatted: "Rp 1.000.000.000",
			PremiumFormatted:    "Rp 2.760.000",
			PaymentFrequency:    "tahun",
			PortalURL:           "http://localhost:3000/portal/status/APP-2026-8819",
		})
		if err != nil {
			t.Fatalf("RenderApplicationSubmitted() error = %v", err)
		}
		if !strings.Contains(text, "APP-2026-8819") || !strings.Contains(html, "APP-2026-8819") {
			t.Errorf("expected ApplicationID in rendered output")
		}
		if !strings.Contains(html, "Rp 1.000.000.000") || !strings.Contains(text, "Rp 1.000.000.000") {
			t.Errorf("expected SumAssuredFormatted in rendered output")
		}
		if !strings.Contains(html, "Bayu Insurance") {
			t.Errorf("expected brand header in html")
		}
	})

	t.Run("ApplicationApproved", func(t *testing.T) {
		text, html, err := renderer.RenderApplicationApproved(ApplicationApprovedData{
			FullName:            "Bayu Pratama",
			PolicyNumber:        "POL-2026-8819",
			ProductName:         "Secure Life Plus",
			ApplicationID:       "APP-2026-8819",
			SumAssuredFormatted: "Rp 1.000.000.000",
			ProtectionPeriod:    "07 Sep 2026 – 07 Sep 2036 (10 Tahun)",
			PremiumFormatted:    "Rp 2.760.000 / tahun",
			PolicyDownloadURL:   "http://localhost:3000/download/POL-2026-8819",
			PortalURL:           "http://localhost:3000/portal/policy/POL-2026-8819",
		})
		if err != nil {
			t.Fatalf("RenderApplicationApproved() error = %v", err)
		}
		if !strings.Contains(text, "POL-2026-8819") || !strings.Contains(html, "POL-2026-8819") {
			t.Errorf("expected PolicyNumber in rendered output")
		}
		if !strings.Contains(html, "Free Look Period") || !strings.Contains(text, "FREE LOOK PERIOD") {
			t.Errorf("expected Free Look Period in rendered output")
		}
	})

	t.Run("ApplicationRejected", func(t *testing.T) {
		text, html, err := renderer.RenderApplicationRejected(ApplicationRejectedData{
			FullName:              "Bayu Pratama",
			ProductName:           "Secure Life Plus",
			ApplicationID:         "APP-2026-8819",
			RejectionCode:         "UW-DEC-401",
			RejectionReason:       "Riwayat kesehatan melampaui batas toleransi risiko produk",
			LeadUnderwriterName:   "dr. Hendra Kurniawan, Sp.Ok",
			LeadUnderwriterNIP:    "UW-2026-042",
			RefundAmountFormatted: "Rp 2.760.000",
			ConsultationURL:       "http://localhost:3000/consultation/APP-2026-8819",
		})
		if err != nil {
			t.Fatalf("RenderApplicationRejected() error = %v", err)
		}
		if !strings.Contains(text, "UW-DEC-401") || !strings.Contains(html, "UW-DEC-401") {
			t.Errorf("expected RejectionCode in rendered output")
		}
		if !strings.Contains(html, "Jaminan Pengembalian Dana") {
			t.Errorf("expected Refund guarantee in html")
		}
	})

	t.Run("ApplicationRFI", func(t *testing.T) {
		text, html, err := renderer.RenderApplicationRFI(ApplicationRFIData{
			FullName:      "Bayu Pratama",
			ProductName:   "Secure Life Plus",
			ApplicationID: "APP-2026-8819",
			Notes:         "Mohon unggah ulang foto e-KTP dan slip gaji 3 bulan terakhir",
			RequiredDocs: []string{
				"1. Foto Ulang Fisik e-KTP (Resolusi Tinggi & Tanpa Pantulan)",
				"2. Slip Gaji 3 Bulan Terakhir / Rekening Koran Legalisir Bank",
			},
			SLADeadline:     "3 x 24 Jam (Maks. 09 Sep 2026, 23:59 WIB)",
			UploadPortalURL: "http://localhost:3000/portal/rfi/APP-2026-8819",
		})
		if err != nil {
			t.Fatalf("RenderApplicationRFI() error = %v", err)
		}
		if !strings.Contains(text, "Foto Ulang Fisik e-KTP") || !strings.Contains(html, "Foto Ulang Fisik e-KTP") {
			t.Errorf("expected RequiredDocs in rendered output")
		}
		if !strings.Contains(html, "3 x 24 Jam") {
			t.Errorf("expected SLADeadline in html")
		}
	})
}

func TestFormatIDR(t *testing.T) {
	cases := []struct {
		input    int64
		expected string
	}{
		{0, "Rp 0"},
		{500, "Rp 500"},
		{1000, "Rp 1.000"},
		{12500, "Rp 12.500"},
		{2760000, "Rp 2.760.000"},
		{1000000000, "Rp 1.000.000.000"},
		{-50000, "-Rp 50.000"},
		{math.MinInt64, "-Rp 9.223.372.036.854.775.808"},
		{math.MaxInt64, "Rp 9.223.372.036.854.775.807"},
	}

	for _, c := range cases {
		actual := FormatIDR(c.input)
		if actual != c.expected {
			t.Errorf("FormatIDR(%d) = %s, want %s", c.input, actual, c.expected)
		}
	}
}
