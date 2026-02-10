package mrz

import "testing"

func TestParseTD3(t *testing.T) {
	raw := "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\nL898902C36UTO7408122F1204159ZE184226B<<<<<10"
	data, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if data.Format != FormatTD3 {
		t.Fatalf("expected format TD3, got %s", data.Format)
	}
	if data.DocumentNumber != "L898902C3" {
		t.Fatalf("unexpected document number: %s", data.DocumentNumber)
	}
	if data.Surname != "ERIKSSON" {
		t.Fatalf("unexpected surname: %s", data.Surname)
	}
	if !data.IsValid {
		t.Fatalf("expected valid MRZ")
	}
}

func TestParseTD1(t *testing.T) {
	raw := "I<UTOD231458907<<<<<<<<<<<<<<<\n7408122F1204159UTO<<<<<<<<<<<0\nERIKSSON<<ANNA<MARIA<<<<<<<<<<"
	data, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if data.Format != FormatTD1 {
		t.Fatalf("expected format TD1, got %s", data.Format)
	}
	if data.DocumentNumber != "D23145890" {
		t.Fatalf("unexpected document number: %s", data.DocumentNumber)
	}
	if data.GivenNames != "ANNA MARIA" {
		t.Fatalf("unexpected given names: %s", data.GivenNames)
	}
	if !data.IsValid {
		t.Fatalf("expected valid MRZ")
	}
}

func TestParseTD2(t *testing.T) {
	raw := "I<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<\nD231458907UTO7408122F1204159<<<<<<<6"
	data, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if data.Format != FormatTD2 {
		t.Fatalf("expected format TD2, got %s", data.Format)
	}
	if data.DocumentNumber != "D23145890" {
		t.Fatalf("unexpected document number: %s", data.DocumentNumber)
	}
	if data.Surname != "ERIKSSON" {
		t.Fatalf("unexpected surname: %s", data.Surname)
	}
	if !data.IsValid {
		t.Fatalf("expected valid MRZ")
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("BAD"); err == nil {
		t.Fatalf("expected error for invalid MRZ")
	}
}

func TestParseMRVA(t *testing.T) {
	raw := "V<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\nL898902C36UTO7408122F1204159ZE184226B<<<<<10"
	data, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if data.Format != FormatMRVA {
		t.Fatalf("expected format MRVA, got %s", data.Format)
	}
	if !data.IsValid {
		t.Fatalf("expected valid MRZ")
	}
}

func TestParseMRVB(t *testing.T) {
	raw := "V<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<\nD231458907UTO7408122F1204159<<<<<<<6"
	data, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if data.Format != FormatMRVB {
		t.Fatalf("expected format MRVB, got %s", data.Format)
	}
	if !data.IsValid {
		t.Fatalf("expected valid MRZ")
	}
}

func TestParseInvalidDate(t *testing.T) {
	if _, err := parseDate("991332"); err == nil {
		t.Fatalf("expected error for invalid date")
	}
}
