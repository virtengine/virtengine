// Package types contains tests for the Fraud module message types.
//
// VE-912: Fraud reporting flow - Message tests
package types

import (
	"strings"
	"testing"
)

func TestMsgSubmitFraudReport_ValidateBasic(t *testing.T) {
	validEvidence := []EncryptedEvidence{{
		AlgorithmID:     "X25519",
		RecipientKeyIDs: []string{"key1"},
		Nonce:           []byte("nonce"),
		Ciphertext:      []byte("data"),
		SenderPubKey:    []byte("pubkey"),
		EvidenceHash:    "hash",
	}}
	validDescription := "This is a description that exceeds the minimum length requirement."

	tests := []struct {
		name    string
		msg     *MsgSubmitFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportedParty: "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
				Category:      FraudCategoryFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: false,
		},
		{
			name: "invalid reporter address",
			msg: &MsgSubmitFraudReport{
				Reporter:      "invalid",
				ReportedParty: "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
				Category:      FraudCategoryFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "invalid reported party address",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportedParty: "invalid",
				Category:      FraudCategoryFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "self report",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportedParty: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				Category:      FraudCategoryFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "invalid category",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportedParty: "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
				Category:      FraudCategoryUnspecified,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "description too short",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportedParty: "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
				Category:      FraudCategoryFakeIdentity,
				Description:   "short",
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "description too long",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportedParty: "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
				Category:      FraudCategoryFakeIdentity,
				Description:   strings.Repeat("a", MaxDescriptionLength+1),
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "missing evidence",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportedParty: "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
				Category:      FraudCategoryFakeIdentity,
				Description:   validDescription,
				Evidence:      []EncryptedEvidence{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgSubmitFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgSubmitFraudReport_GetSigners(t *testing.T) {
	msg := &MsgSubmitFraudReport{
		Reporter: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
	}

	signers := msg.GetSigners()
	if len(signers) != 1 {
		t.Errorf("GetSigners() returned %d signers, want 1", len(signers))
	}
}

func TestMsgSubmitFraudReport_Route(t *testing.T) {
	msg := &MsgSubmitFraudReport{}
	if got := msg.Route(); got != RouterKey {
		t.Errorf("Route() = %v, want %v", got, RouterKey)
	}
}

func TestMsgSubmitFraudReport_Type(t *testing.T) {
	msg := &MsgSubmitFraudReport{}
	if got := msg.Type(); got != TypeMsgSubmitFraudReport {
		t.Errorf("Type() = %v, want %v", got, TypeMsgSubmitFraudReport)
	}
}

func TestMsgAssignModerator_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgAssignModerator
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgAssignModerator{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				AssignTo:  "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgAssignModerator{
				Moderator: "invalid",
				ReportID:  "fraud-report-1",
				AssignTo:  "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgAssignModerator{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "",
				AssignTo:  "cosmos1qqqqszqgpqyqszqgpqyqszqgpqyqszqgpax7d8",
			},
			wantErr: true,
		},
		{
			name: "invalid assign_to address",
			msg: &MsgAssignModerator{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				AssignTo:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgAssignModerator.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgUpdateReportStatus_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgUpdateReportStatus
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgUpdateReportStatus{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				NewStatus: FraudReportStatusReviewing,
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgUpdateReportStatus{
				Moderator: "invalid",
				ReportID:  "fraud-report-1",
				NewStatus: FraudReportStatusReviewing,
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgUpdateReportStatus{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "",
				NewStatus: FraudReportStatusReviewing,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			msg: &MsgUpdateReportStatus{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				NewStatus: FraudReportStatusUnspecified,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgUpdateReportStatus.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgResolveFraudReport_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgResolveFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgResolveFraudReport{
				Moderator:  "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:   "fraud-report-1",
				Resolution: ResolutionTypeWarning,
				Notes:      "User has been warned",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgResolveFraudReport{
				Moderator:  "invalid",
				ReportID:   "fraud-report-1",
				Resolution: ResolutionTypeWarning,
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgResolveFraudReport{
				Moderator:  "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:   "",
				Resolution: ResolutionTypeWarning,
			},
			wantErr: true,
		},
		{
			name: "invalid resolution",
			msg: &MsgResolveFraudReport{
				Moderator:  "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:   "fraud-report-1",
				Resolution: ResolutionTypeUnspecified,
			},
			wantErr: true,
		},
		{
			name: "notes too long",
			msg: &MsgResolveFraudReport{
				Moderator:  "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:   "fraud-report-1",
				Resolution: ResolutionTypeWarning,
				Notes:      strings.Repeat("a", MaxResolutionNotesLength+1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgResolveFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgRejectFraudReport_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgRejectFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgRejectFraudReport{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				Notes:     "No evidence of fraud",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgRejectFraudReport{
				Moderator: "invalid",
				ReportID:  "fraud-report-1",
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgRejectFraudReport{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "",
			},
			wantErr: true,
		},
		{
			name: "notes too long",
			msg: &MsgRejectFraudReport{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				Notes:     strings.Repeat("a", MaxResolutionNotesLength+1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgRejectFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgEscalateFraudReport_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgEscalateFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgEscalateFraudReport{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				Reason:    "Requires admin review",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgEscalateFraudReport{
				Moderator: "invalid",
				ReportID:  "fraud-report-1",
				Reason:    "Requires admin review",
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgEscalateFraudReport{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "",
				Reason:    "Requires admin review",
			},
			wantErr: true,
		},
		{
			name: "empty reason",
			msg: &MsgEscalateFraudReport{
				Moderator: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				ReportID:  "fraud-report-1",
				Reason:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgEscalateFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgUpdateParams
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgUpdateParams{
				Authority: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				Params:    DefaultParams(),
			},
			wantErr: false,
		},
		{
			name: "invalid authority address",
			msg: &MsgUpdateParams{
				Authority: "invalid",
				Params:    DefaultParams(),
			},
			wantErr: true,
		},
		{
			name: "invalid params",
			msg: &MsgUpdateParams{
				Authority: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp5xf2c9",
				Params: Params{
					MinDescriptionLength: 5, // Invalid
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgUpdateParams.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMsgSubmitFraudReport(t *testing.T) {
	evidence := []EncryptedEvidence{{
		AlgorithmID:     "X25519",
		RecipientKeyIDs: []string{"key1"},
		Nonce:           []byte("nonce"),
		Ciphertext:      []byte("data"),
		SenderPubKey:    []byte("pubkey"),
		EvidenceHash:    "hash",
	}}

	msg := NewMsgSubmitFraudReport(
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"This is a valid description",
		evidence,
		[]string{"order-1", "order-2"},
	)

	if msg.Reporter != "cosmos1reporter" {
		t.Errorf("NewMsgSubmitFraudReport() Reporter = %v, want %v", msg.Reporter, "cosmos1reporter")
	}
	if msg.Category != FraudCategoryFakeIdentity {
		t.Errorf("NewMsgSubmitFraudReport() Category = %v, want %v", msg.Category, FraudCategoryFakeIdentity)
	}
	if len(msg.RelatedOrderIDs) != 2 {
		t.Errorf("NewMsgSubmitFraudReport() RelatedOrderIDs length = %v, want %v", len(msg.RelatedOrderIDs), 2)
	}
}

func TestNewMsgResolveFraudReport(t *testing.T) {
	msg := NewMsgResolveFraudReport(
		"cosmos1moderator",
		"fraud-report-1",
		ResolutionTypeWarning,
		"User warned",
	)

	if msg.Moderator != "cosmos1moderator" {
		t.Errorf("NewMsgResolveFraudReport() Moderator = %v, want %v", msg.Moderator, "cosmos1moderator")
	}
	if msg.Resolution != ResolutionTypeWarning {
		t.Errorf("NewMsgResolveFraudReport() Resolution = %v, want %v", msg.Resolution, ResolutionTypeWarning)
	}
}
