package marketplace

import "testing"

func TestParseOrderID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: "cosmos1abc/1", wantErr: false},
		{name: "missing_parts", value: "cosmos1abc", wantErr: true},
		{name: "bad_sequence", value: "cosmos1abc/notanum", wantErr: true},
		{name: "zero_sequence", value: "cosmos1abc/0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseOrderID(tt.value)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %s", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.value, err)
			}
		})
	}
}

func TestParseAllocationID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: "cosmos1abc/1/2", wantErr: false},
		{name: "missing_parts", value: "cosmos1abc/1", wantErr: true},
		{name: "bad_order_sequence", value: "cosmos1abc/notanum/2", wantErr: true},
		{name: "bad_allocation_sequence", value: "cosmos1abc/1/notanum", wantErr: true},
		{name: "zero_order_sequence", value: "cosmos1abc/0/1", wantErr: true},
		{name: "zero_allocation_sequence", value: "cosmos1abc/1/0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAllocationID(tt.value)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %s", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.value, err)
			}
		})
	}
}

func TestParseOfferingID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: "cosmos1abc/1", wantErr: false},
		{name: "missing_parts", value: "cosmos1abc", wantErr: true},
		{name: "bad_sequence", value: "cosmos1abc/notanum", wantErr: true},
		{name: "zero_sequence", value: "cosmos1abc/0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseOfferingID(tt.value)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %s", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.value, err)
			}
		})
	}
}
