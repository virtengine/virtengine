# Task 31F: International ID Document Adapters

**vibe-kanban ID:** `9d7f3d3b-f2f2-4419-8614-75c3a19e2a21`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31F |
| **Title** | feat(veid): International ID document adapters |
| **Priority** | P2 |
| **Wave** | 4 |
| **Estimated LOC** | 4500 |
| **Duration** | 4-5 weeks |
| **Dependencies** | VEID ML pipeline |
| **Blocking** | None |

---

## Problem Statement

The patent specification mentions support for multiple international ID document formats. Currently, the VEID system only supports a limited set of document types. For global adoption, we need:

- European ID card support (various formats)
- Asian passport formats
- Middle Eastern ID formats  
- African ID document support
- South American ID documents
- Machine Readable Zone (MRZ) parsing
- NFC passport reading (ePassport)
- Multi-language OCR support

### Current State Analysis

```
ml/veid_pipeline/ocr/           ⚠️  Limited document types
ml/veid_pipeline/ocr/mrz/       ❌  Does not exist
pkg/veid/documents/             ❌  No adapter pattern
lib/capture/                    ⚠️  Basic document capture
```

---

## Acceptance Criteria

### AC-1: Document Format Adapters
- [ ] Create adapter interface for ID document types
- [ ] European ID cards (EU-wide standard)
- [ ] UK passport and driver's license
- [ ] US passport and driver's license (all states)
- [ ] Canadian passport and ID
- [ ] Australian passport and driver's license
- [ ] Japanese ID cards (My Number Card)
- [ ] South Korean ID (resident registration)
- [ ] Indian Aadhaar card support
- [ ] UAE ID card support
- [ ] Brazilian CPF/CNH support

### AC-2: MRZ Parsing
- [ ] TD1 format (ID cards, 3 lines × 30 chars)
- [ ] TD2 format (travel documents, 2 lines × 36 chars)
- [ ] TD3 format (passports, 2 lines × 44 chars)
- [ ] MRV-A and MRV-B (visas)
- [ ] Check digit validation
- [ ] Composite check digit verification
- [ ] Date parsing (YYMMDD → actual date)
- [ ] Country code mapping

### AC-3: Multi-Language OCR
- [ ] Latin script optimization
- [ ] Cyrillic script support (Russian IDs)
- [ ] Arabic script support (RTL handling)
- [ ] CJK character support (Chinese, Japanese, Korean)
- [ ] Devanagari script (Hindi IDs)
- [ ] Thai script support
- [ ] Language detection and auto-routing

### AC-4: NFC ePassport Reading
- [ ] PACE protocol implementation
- [ ] BAC (Basic Access Control) support
- [ ] EAC (Extended Access Control) for EU passports
- [ ] DG1 (MRZ data) extraction
- [ ] DG2 (facial image) extraction
- [ ] Passive authentication (SOD verification)
- [ ] Active authentication (chip genuineness)

---

## Technical Requirements

### Document Adapter Interface

```go
// pkg/veid/documents/adapter.go

package documents

import (
    "context"
    "image"
)

// DocumentType identifies the document category
type DocumentType string

const (
    DocumentTypePassport       DocumentType = "passport"
    DocumentTypeIDCard         DocumentType = "id_card"
    DocumentTypeDriverLicense  DocumentType = "driver_license"
    DocumentTypeResidencePermit DocumentType = "residence_permit"
)

// CountryCode is ISO 3166-1 alpha-3
type CountryCode string

// DocumentAdapter extracts data from a specific document format
type DocumentAdapter interface {
    // Supported returns the document types this adapter handles
    SupportedTypes() []DocumentType
    
    // SupportedCountries returns the countries this adapter handles
    SupportedCountries() []CountryCode
    
    // CanProcess checks if this adapter can handle the given document
    CanProcess(docType DocumentType, country CountryCode) bool
    
    // Extract extracts structured data from the document image
    Extract(ctx context.Context, img image.Image) (*DocumentData, error)
    
    // ExtractWithMRZ extracts data using both visual and MRZ
    ExtractWithMRZ(ctx context.Context, img image.Image, mrz string) (*DocumentData, error)
    
    // Validate performs document-specific validation
    Validate(data *DocumentData) ([]ValidationError, error)
}

// DocumentData is the standardized output from any adapter
type DocumentData struct {
    // Core identity fields
    GivenNames    string
    Surname       string
    DateOfBirth   time.Time
    Sex           string  // M, F, X
    Nationality   CountryCode
    
    // Document fields
    DocumentType   DocumentType
    DocumentNumber string
    IssuingCountry CountryCode
    ExpiryDate     time.Time
    IssueDate      *time.Time  // Optional
    
    // Additional data
    PlaceOfBirth   string
    Address        *Address
    
    // Biometric data
    FacialImage    []byte  // JPEG extracted face
    
    // MRZ data if available
    MRZData        *MRZData
    
    // NFC data if available
    NFCData        *NFCData
    
    // Confidence scores
    OverallConfidence float64
    FieldConfidences  map[string]float64
}
```

### MRZ Parser

```go
// pkg/veid/documents/mrz/parser.go

package mrz

import (
    "errors"
    "regexp"
    "time"
)

type MRZFormat string

const (
    FormatTD1  MRZFormat = "TD1"  // ID cards (90 chars, 3 lines)
    FormatTD2  MRZFormat = "TD2"  // Travel docs (72 chars, 2 lines)
    FormatTD3  MRZFormat = "TD3"  // Passports (88 chars, 2 lines)
    FormatMRVA MRZFormat = "MRVA" // Visa Type A
    FormatMRVB MRZFormat = "MRVB" // Visa Type B
)

type MRZData struct {
    Raw            string
    Format         MRZFormat
    DocumentType   string  // P, I, A, C, etc.
    IssuingCountry string  // 3-letter
    Surname        string
    GivenNames     string
    DocumentNumber string
    Nationality    string
    DateOfBirth    time.Time
    Sex            string
    ExpiryDate     time.Time
    OptionalData1  string
    OptionalData2  string
    
    // Validation
    CheckDigits    CheckDigits
    IsValid        bool
}

type CheckDigits struct {
    DocumentNumber  int
    DateOfBirth     int
    ExpiryDate      int
    OptionalData    int
    Composite       int
    
    DocumentNumberValid  bool
    DateOfBirthValid     bool
    ExpiryDateValid      bool
    OptionalDataValid    bool
    CompositeValid       bool
}

func Parse(mrz string) (*MRZData, error) {
    // Clean and normalize
    mrz = normalizeMRZ(mrz)
    
    // Detect format
    format := detectFormat(mrz)
    if format == "" {
        return nil, errors.New("unrecognized MRZ format")
    }
    
    switch format {
    case FormatTD1:
        return parseTD1(mrz)
    case FormatTD2:
        return parseTD2(mrz)
    case FormatTD3:
        return parseTD3(mrz)
    default:
        return nil, errors.New("unsupported MRZ format")
    }
}

func parseTD3(mrz string) (*MRZData, error) {
    if len(mrz) != 88 {
        return nil, errors.New("TD3 must be 88 characters")
    }
    
    line1 := mrz[0:44]
    line2 := mrz[44:88]
    
    data := &MRZData{
        Raw:            mrz,
        Format:         FormatTD3,
        DocumentType:   string(line1[0]),
        IssuingCountry: string(line1[2:5]),
        Surname:        extractName(line1[5:44], true),
        GivenNames:     extractName(line1[5:44], false),
        DocumentNumber: string(line2[0:9]),
        Nationality:    string(line2[10:13]),
        Sex:            string(line2[20]),
        OptionalData1:  trim(line2[28:42]),
    }
    
    // Parse dates
    data.DateOfBirth = parseMRZDate(string(line2[13:19]))
    data.ExpiryDate = parseMRZDate(string(line2[21:27]))
    
    // Validate check digits
    data.CheckDigits = validateTD3CheckDigits(mrz)
    data.IsValid = data.CheckDigits.CompositeValid
    
    return data, nil
}

// Check digit calculation per ICAO 9303
func calculateCheckDigit(s string) int {
    weights := []int{7, 3, 1}
    sum := 0
    
    for i, c := range s {
        var value int
        switch {
        case c >= '0' && c <= '9':
            value = int(c - '0')
        case c >= 'A' && c <= 'Z':
            value = int(c - 'A' + 10)
        case c == '<':
            value = 0
        default:
            value = 0
        }
        sum += value * weights[i%3]
    }
    
    return sum % 10
}
```

### NFC ePassport Reader

```go
// pkg/veid/documents/nfc/reader.go

package nfc

import (
    "context"
    "crypto/cipher"
    "errors"
)

// NFCReader handles ePassport NFC communication
type NFCReader interface {
    // Connect establishes connection to the NFC chip
    Connect(ctx context.Context) error
    
    // PerformBAC establishes Basic Access Control session
    PerformBAC(ctx context.Context, mrzInfo MRZInfo) error
    
    // PerformPACE establishes PACE session (modern passports)
    PerformPACE(ctx context.Context, accessKey []byte) error
    
    // ReadDG reads a data group from the chip
    ReadDG(ctx context.Context, dgNum int) ([]byte, error)
    
    // VerifyPassiveAuth verifies the Document Security Object
    VerifyPassiveAuth(ctx context.Context) error
    
    // PerformActiveAuth verifies chip genuineness
    PerformActiveAuth(ctx context.Context) error
    
    // Close terminates the NFC session
    Close() error
}

// MRZInfo contains data needed for BAC
type MRZInfo struct {
    DocumentNumber string
    DateOfBirth    string // YYMMDD
    ExpiryDate     string // YYMMDD
}

// NFCData contains extracted NFC data
type NFCData struct {
    // Data Groups
    DG1  *DG1Data  // MRZ data
    DG2  *DG2Data  // Facial image
    DG3  *DG3Data  // Fingerprints (if accessible)
    DG7  *DG7Data  // Signature (if accessible)
    DG11 *DG11Data // Additional personal data
    DG12 *DG12Data // Additional document data
    
    // Security
    SOD           *SecurityObject
    PassiveAuth   bool
    ActiveAuth    bool
    ChipCloned    bool  // Detected clone attempt
}

// DG1Data is MRZ from NFC
type DG1Data struct {
    MRZData *MRZData
}

// DG2Data is facial image from NFC
type DG2Data struct {
    Images []BiometricImage
}

type BiometricImage struct {
    Type     string // "face", "finger", "iris"
    Format   string // "JPEG", "JPEG2000", "WSQ"
    Data     []byte
    Quality  int
}

// Implementation of BAC protocol
func performBAC(reader NFCReader, mrz MRZInfo) (*BACSession, error) {
    // 1. Derive keys from MRZ
    kEnc, kMac := deriveKeys(mrz)
    
    // 2. Get challenge from chip
    challenge, err := reader.getChallenge()
    if err != nil {
        return nil, err
    }
    
    // 3. Generate and encrypt response
    response := generateBACResponse(challenge, kEnc)
    
    // 4. Mutual authentication
    chipResponse, err := reader.mutualAuthenticate(response)
    if err != nil {
        return nil, errors.New("BAC authentication failed")
    }
    
    // 5. Derive session keys
    return deriveSessionKeys(kEnc, kMac, chipResponse)
}

// Key derivation from MRZ data
func deriveKeys(mrz MRZInfo) (kEnc, kMac []byte) {
    // K_seed = SHA1(MRZ_information)
    seedData := mrz.DocumentNumber + 
                calculateCheckDigit(mrz.DocumentNumber) +
                mrz.DateOfBirth +
                calculateCheckDigit(mrz.DateOfBirth) +
                mrz.ExpiryDate +
                calculateCheckDigit(mrz.ExpiryDate)
    
    kSeed := sha1.Sum([]byte(seedData))
    
    // Derive encryption and MAC keys
    kEnc = deriveKey(kSeed[:16], []byte{0x00, 0x00, 0x00, 0x01})
    kMac = deriveKey(kSeed[:16], []byte{0x00, 0x00, 0x00, 0x02})
    
    return kEnc, kMac
}
```

### Multi-Language OCR Configuration

```python
# ml/veid_pipeline/ocr/multilang.py

from typing import Dict, List, Optional, Tuple
import cv2
import numpy as np
from enum import Enum

class Script(Enum):
    LATIN = "latin"
    CYRILLIC = "cyrillic"
    ARABIC = "arabic"
    CJK = "cjk"
    DEVANAGARI = "devanagari"
    THAI = "thai"
    HEBREW = "hebrew"
    GREEK = "greek"

# Language to script mapping
LANG_SCRIPT_MAP: Dict[str, Script] = {
    "en": Script.LATIN,
    "de": Script.LATIN,
    "fr": Script.LATIN,
    "es": Script.LATIN,
    "it": Script.LATIN,
    "pt": Script.LATIN,
    "ru": Script.CYRILLIC,
    "uk": Script.CYRILLIC,
    "ar": Script.ARABIC,
    "fa": Script.ARABIC,
    "zh": Script.CJK,
    "ja": Script.CJK,
    "ko": Script.CJK,
    "hi": Script.DEVANAGARI,
    "th": Script.THAI,
    "he": Script.HEBREW,
    "el": Script.GREEK,
}

# Tesseract language codes
TESSERACT_LANG_MAP: Dict[str, str] = {
    "en": "eng",
    "de": "deu",
    "fr": "fra",
    "es": "spa",
    "ru": "rus",
    "ar": "ara",
    "zh": "chi_sim+chi_tra",
    "ja": "jpn",
    "ko": "kor",
    "hi": "hin",
    "th": "tha",
}

class MultiLangOCR:
    """Multi-language OCR with script detection and routing"""
    
    def __init__(self, config: Optional[Dict] = None):
        self.config = config or {}
        self._init_models()
    
    def _init_models(self):
        """Initialize OCR models for different scripts"""
        self.script_detector = ScriptDetector()
        self.ocr_engines = {}
        
        # Initialize engines lazily
        for script in Script:
            self.ocr_engines[script] = None
    
    def detect_script(self, image: np.ndarray) -> Tuple[Script, float]:
        """Detect the dominant script in an image"""
        return self.script_detector.detect(image)
    
    def extract_text(
        self,
        image: np.ndarray,
        language: Optional[str] = None,
        regions: Optional[List[Tuple[int, int, int, int]]] = None,
    ) -> Dict[str, any]:
        """Extract text from image with auto language detection"""
        
        # Detect script if language not specified
        if language is None:
            script, confidence = self.detect_script(image)
            language = self._get_default_language(script)
        else:
            script = LANG_SCRIPT_MAP.get(language, Script.LATIN)
        
        # Get appropriate OCR engine
        engine = self._get_engine(script)
        
        # Handle RTL scripts
        if script == Script.ARABIC or script == Script.HEBREW:
            image = self._preprocess_rtl(image)
        
        # Run OCR
        if regions:
            results = []
            for x, y, w, h in regions:
                roi = image[y:y+h, x:x+w]
                text = engine.recognize(roi, language)
                results.append({
                    "text": text,
                    "bbox": (x, y, w, h),
                    "confidence": engine.confidence,
                })
            return {"regions": results}
        else:
            text = engine.recognize(image, language)
            return {
                "text": text,
                "confidence": engine.confidence,
                "script": script.value,
                "language": language,
            }
    
    def _get_engine(self, script: Script):
        """Get or initialize OCR engine for script"""
        if self.ocr_engines[script] is None:
            self.ocr_engines[script] = self._create_engine(script)
        return self.ocr_engines[script]
    
    def _create_engine(self, script: Script):
        """Create OCR engine optimized for script"""
        if script == Script.CJK:
            return CJKOCREngine()
        elif script == Script.ARABIC:
            return ArabicOCREngine()
        else:
            return TesseractOCREngine(script)
    
    def _preprocess_rtl(self, image: np.ndarray) -> np.ndarray:
        """Preprocess image for RTL text"""
        # For RTL, sometimes flipping helps with reading order
        # But modern OCR handles this automatically
        return image
```

---

## Directory Structure

```
pkg/veid/documents/
├── adapter.go            # Adapter interface
├── registry.go           # Adapter registry
├── types.go              # Common types
├── passport/
│   ├── generic.go        # Generic passport adapter
│   ├── us.go             # US passport specifics
│   ├── uk.go             # UK passport specifics
│   ├── eu.go             # EU passport specifics
│   └── jp.go             # Japanese passport
├── idcard/
│   ├── generic.go        # Generic ID card
│   ├── eu.go             # EU ID card
│   ├── aadhaar.go        # Indian Aadhaar
│   └── uae.go            # UAE ID
├── mrz/
│   ├── parser.go         # MRZ parsing
│   ├── validator.go      # Check digit validation
│   └── formats.go        # Format definitions
└── nfc/
    ├── reader.go         # NFC interface
    ├── bac.go            # Basic Access Control
    ├── pace.go           # PACE protocol
    ├── datagroups.go     # DG parsing
    └── security.go       # Passive/Active auth

ml/veid_pipeline/ocr/
├── multilang.py          # Multi-language OCR
├── scripts/
│   ├── latin.py
│   ├── cyrillic.py
│   ├── arabic.py
│   ├── cjk.py
│   └── devanagari.py
├── mrz/
│   ├── detector.py       # MRZ region detection
│   └── reader.py         # MRZ OCR
└── models/
    └── (script-specific models)
```

---

## Testing Requirements

### Unit Tests
- MRZ parsing for all formats
- Check digit validation
- Country code mapping

### Integration Tests
- Full document extraction pipeline
- Multi-language OCR accuracy
- NFC reading (with emulator)

### Validation Tests
- Test with sample documents from each country
- Edge cases (damaged, expired, etc.)
- Anti-fraud detection

---

## Security Considerations

1. **Document Images**: Never persist raw document images
2. **NFC Keys**: MRZ-derived keys used only in memory
3. **PII Handling**: Encrypt extracted data immediately
4. **Anti-Cloning**: Detect NFC chip cloning attempts
5. **Audit Trail**: Log all document processing events
