package types

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const (
	sgxQuoteHeaderSize            = 48
	sgxReportBodySize             = 384
	sgxQuoteSigLenOffset          = sgxQuoteHeaderSize + sgxReportBodySize
	sgxQuoteSigDataMinSize        = 64 + 64 + 384 + 64 + 2 + 2 + 4
	sgxQuoteSigDataOffset         = sgxQuoteSigLenOffset + 4
	sgxReportBodyOffset           = sgxQuoteHeaderSize
	sgxReportAttributesOff        = 48
	sgxReportMRENCLAVEOff         = 64
	sgxReportMRSIGNEROff          = 128
	sgxReportISVProdIDOff         = 256
	sgxReportISVSVNOff            = 258
	sgxReportReportDataOff        = 320
	sgxAttrDebugFlag       uint64 = 0x02
)

const (
	sevSnpMinReportSize            = 0x4A0 // 1184 bytes
	sevSnpVersionOffset            = 0x000
	sevSnpPolicyOffset             = 0x008
	sevSnpReportDataOffset         = 0x050
	sevSnpMeasurementOffset        = 0x090
	sevSnpIDKeyDigestOffset        = 0x0E0
	sevSnpAuthorKeyOffset          = 0x100
	sevSnpCurrentTCBOffset         = 0x038
	sevSnpReportedTCBOffset        = 0x160
	sevSnpCommittedTCBOff          = 0x1A0
	sevSnpDebugPolicy       uint64 = 1 << 19
)

// SGXQuoteHeader is the DCAP quote header (v3+).
type SGXQuoteHeader struct {
	Version            uint16
	AttestationKeyType uint16
	TEEType            uint32
	QESVN              uint16
	PCESVN             uint16
	QEVendorID         [16]byte
	UserData           [20]byte
}

// SGXReportBody captures the SGX report body fields we need for validation.
type SGXReportBody struct {
	CPUSVN     [16]byte
	Attributes [16]byte
	MRENCLAVE  [32]byte
	MRSIGNER   [32]byte
	ISVProdID  uint16
	ISVSVN     uint16
	ReportData [64]byte
}

// DebugEnabled returns true if the report attributes indicate a debug enclave.
func (r SGXReportBody) DebugEnabled() bool {
	flags := binary.LittleEndian.Uint64(r.Attributes[0:8])
	return flags&sgxAttrDebugFlag != 0
}

// SGXDCAPQuote represents the parsed SGX DCAP quote.
type SGXDCAPQuote struct {
	Header   SGXQuoteHeader
	Report   SGXReportBody
	QEReport SGXReportBody
}

// ParseSGXDCAPQuoteV3 parses a DCAP v3 quote and extracts report data.
func ParseSGXDCAPQuoteV3(quote []byte) (*SGXDCAPQuote, error) {
	if len(quote) < sgxQuoteSigDataOffset+sgxQuoteSigDataMinSize {
		return nil, fmt.Errorf("quote too small: got %d bytes, need at least %d", len(quote), sgxQuoteSigDataOffset+sgxQuoteSigDataMinSize)
	}

	header := SGXQuoteHeader{
		Version:            binary.LittleEndian.Uint16(quote[0:2]),
		AttestationKeyType: binary.LittleEndian.Uint16(quote[2:4]),
		TEEType:            binary.LittleEndian.Uint32(quote[4:8]),
		QESVN:              binary.LittleEndian.Uint16(quote[8:10]),
		PCESVN:             binary.LittleEndian.Uint16(quote[10:12]),
	}
	copy(header.QEVendorID[:], quote[12:28])
	copy(header.UserData[:], quote[28:48])

	report, err := parseSGXReportBody(quote[sgxReportBodyOffset : sgxReportBodyOffset+sgxReportBodySize])
	if err != nil {
		return nil, fmt.Errorf("parse report body: %w", err)
	}

	// Signature data layout:
	// [ISV Sig 64][AttKey 64][QE Report 384][QE Sig 64][AuthSize 2][AuthData ...][CertType 2][CertSize 4][CertData ...]
	sigData := quote[sgxQuoteSigDataOffset:]
	qeReportOffset := 64 + 64
	qeReportBytes := sigData[qeReportOffset : qeReportOffset+sgxReportBodySize]
	qeReport, err := parseSGXReportBody(qeReportBytes)
	if err != nil {
		return nil, fmt.Errorf("parse QE report body: %w", err)
	}

	return &SGXDCAPQuote{
		Header:   header,
		Report:   report,
		QEReport: qeReport,
	}, nil
}

func parseSGXReportBody(report []byte) (SGXReportBody, error) {
	if len(report) < sgxReportBodySize {
		return SGXReportBody{}, fmt.Errorf("report body too small: got %d bytes, need %d", len(report), sgxReportBodySize)
	}

	var body SGXReportBody
	copy(body.CPUSVN[:], report[0:16])
	copy(body.Attributes[:], report[sgxReportAttributesOff:sgxReportAttributesOff+16])
	copy(body.MRENCLAVE[:], report[sgxReportMRENCLAVEOff:sgxReportMRENCLAVEOff+32])
	copy(body.MRSIGNER[:], report[sgxReportMRSIGNEROff:sgxReportMRSIGNEROff+32])
	body.ISVProdID = binary.LittleEndian.Uint16(report[sgxReportISVProdIDOff : sgxReportISVProdIDOff+2])
	body.ISVSVN = binary.LittleEndian.Uint16(report[sgxReportISVSVNOff : sgxReportISVSVNOff+2])
	copy(body.ReportData[:], report[sgxReportReportDataOff:sgxReportReportDataOff+64])
	return body, nil
}

// SEVSNPReport represents a parsed SEV-SNP attestation report.
type SEVSNPReport struct {
	Version         uint32
	Policy          uint64
	ReportData      [64]byte
	Measurement     [48]byte
	IDKeyDigest     [32]byte
	AuthorKeyDigest [32]byte
	CurrentTCB      uint64
	ReportedTCB     uint64
	CommittedTCB    uint64
}

// DebugEnabled returns true when the SNP policy allows debug mode.
func (r SEVSNPReport) DebugEnabled() bool {
	return r.Policy&sevSnpDebugPolicy != 0
}

// ParseSEVSNPReport parses a SEV-SNP attestation report.
func ParseSEVSNPReport(report []byte) (*SEVSNPReport, error) {
	if len(report) < sevSnpMinReportSize {
		return nil, fmt.Errorf("report too small: got %d bytes, need %d", len(report), sevSnpMinReportSize)
	}

	parsed := &SEVSNPReport{
		Version:      binary.LittleEndian.Uint32(report[sevSnpVersionOffset : sevSnpVersionOffset+4]),
		Policy:       binary.LittleEndian.Uint64(report[sevSnpPolicyOffset : sevSnpPolicyOffset+8]),
		CurrentTCB:   binary.LittleEndian.Uint64(report[sevSnpCurrentTCBOffset : sevSnpCurrentTCBOffset+8]),
		ReportedTCB:  binary.LittleEndian.Uint64(report[sevSnpReportedTCBOffset : sevSnpReportedTCBOffset+8]),
		CommittedTCB: binary.LittleEndian.Uint64(report[sevSnpCommittedTCBOff : sevSnpCommittedTCBOff+8]),
	}
	copy(parsed.ReportData[:], report[sevSnpReportDataOffset:sevSnpReportDataOffset+64])
	copy(parsed.Measurement[:], report[sevSnpMeasurementOffset:sevSnpMeasurementOffset+48])
	copy(parsed.IDKeyDigest[:], report[sevSnpIDKeyDigestOffset:sevSnpIDKeyDigestOffset+32])
	copy(parsed.AuthorKeyDigest[:], report[sevSnpAuthorKeyOffset:sevSnpAuthorKeyOffset+32])

	return parsed, nil
}

// SEVSNPMeasurementHash computes a stable hash for a SEV-SNP measurement.
func SEVSNPMeasurementHash(measurement []byte) [32]byte {
	return sha256.Sum256(measurement)
}
