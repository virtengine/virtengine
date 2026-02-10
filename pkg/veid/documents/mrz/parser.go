package mrz

import (
	"errors"
	"strings"
	"time"
)

type MRZFormat string

const (
	FormatTD1  MRZFormat = "TD1"
	FormatTD2  MRZFormat = "TD2"
	FormatTD3  MRZFormat = "TD3"
	FormatMRVA MRZFormat = "MRVA"
	FormatMRVB MRZFormat = "MRVB"
)

type MRZData struct {
	Raw            string
	Format         MRZFormat
	DocumentType   string
	IssuingCountry string
	Surname        string
	GivenNames     string
	DocumentNumber string
	Nationality    string
	DateOfBirth    time.Time
	Sex            string
	ExpiryDate     time.Time
	OptionalData1  string
	OptionalData2  string
	CheckDigits    CheckDigits
	IsValid        bool
}

type CheckDigits struct {
	DocumentNumber int
	DateOfBirth    int
	ExpiryDate     int
	OptionalData   int
	Composite      int
}

var (
	ErrInvalidLength = errors.New("invalid mrz length")
	ErrInvalidFormat = errors.New("invalid mrz format")
	ErrInvalidDate   = errors.New("invalid mrz date")
)

func Parse(raw string) (*MRZData, error) {
	lines, err := normalizeLines(raw)
	if err != nil {
		return nil, err
	}

	switch {
	case len(lines) == 3 && len(lines[0]) == 30:
		return parseTD1(lines)
	case len(lines) == 2 && len(lines[0]) == 44:
		if strings.HasPrefix(lines[0], "V") {
			return parseTD3(lines, FormatMRVA)
		}
		return parseTD3(lines, FormatTD3)
	case len(lines) == 2 && len(lines[0]) == 36:
		if strings.HasPrefix(lines[0], "V") {
			return parseTD2(lines, FormatMRVB)
		}
		return parseTD2(lines, FormatTD2)
	default:
		return nil, ErrInvalidLength
	}
}

func normalizeLines(raw string) ([]string, error) {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return nil, ErrInvalidLength
	}
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "\t", "")
	cleaned = strings.ToUpper(cleaned)
	parts := strings.Split(cleaned, "\n")

	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			lines = append(lines, part)
		}
	}

	if len(lines) == 1 {
		line := lines[0]
		switch len(line) {
		case 90:
			lines = []string{line[0:30], line[30:60], line[60:90]}
		case 88:
			lines = []string{line[0:44], line[44:88]}
		case 72:
			lines = []string{line[0:36], line[36:72]}
		default:
			return nil, ErrInvalidLength
		}
	}

	return lines, nil
}

func parseTD1(lines []string) (*MRZData, error) {
	if len(lines) != 3 || len(lines[0]) != 30 || len(lines[1]) != 30 || len(lines[2]) != 30 {
		return nil, ErrInvalidFormat
	}

	line1 := lines[0]
	line2 := lines[1]
	line3 := lines[2]

	docType := strings.TrimRight(line1[0:2], "<")
	issuing := line1[2:5]
	documentNumber := line1[5:14]
	documentNumberCheck := line1[14]
	optional1 := line1[15:30]

	birth := line2[0:6]
	birthCheck := line2[6]
	sex := line2[7:8]
	expiry := line2[8:14]
	expiryCheck := line2[14]
	nationality := line2[15:18]
	optional2 := line2[18:29]
	optionalCheck := line2[29]

	surname, given := ParseName(line3)

	data := &MRZData{
		Raw:            strings.Join(lines, "\n"),
		Format:         FormatTD1,
		DocumentType:   docType,
		IssuingCountry: issuing,
		Surname:        surname,
		GivenNames:     given,
		DocumentNumber: Clean(documentNumber),
		Nationality:    nationality,
		Sex:            Clean(sex),
		OptionalData1:  Clean(optional1),
		OptionalData2:  Clean(optional2),
		CheckDigits: CheckDigits{
			DocumentNumber: digitValue(documentNumberCheck),
			DateOfBirth:    digitValue(birthCheck),
			ExpiryDate:     digitValue(expiryCheck),
			OptionalData:   digitValue(optionalCheck),
			Composite:      digitValue(line3[29]),
		},
	}

	var err error
	data.DateOfBirth, err = parseDate(birth)
	if err != nil {
		return nil, err
	}
	data.ExpiryDate, err = parseDate(expiry)
	if err != nil {
		return nil, err
	}

	docOK := checkMatch(documentNumber, data.CheckDigits.DocumentNumber)
	birthOK := checkMatch(birth, data.CheckDigits.DateOfBirth)
	expiryOK := checkMatch(expiry, data.CheckDigits.ExpiryDate)
	optionalOK := checkMatch(optional1+optional2, data.CheckDigits.OptionalData)
	compositeSource := line1[5:30] + line2[0:7] + line2[8:15] + line2[18:29]
	compositeOK := checkMatch(compositeSource, data.CheckDigits.Composite)
	data.IsValid = docOK && birthOK && expiryOK && optionalOK && compositeOK

	return data, nil
}

func parseTD2(lines []string, format MRZFormat) (*MRZData, error) {
	if len(lines) != 2 || len(lines[0]) != 36 || len(lines[1]) != 36 {
		return nil, ErrInvalidFormat
	}

	line1 := lines[0]
	line2 := lines[1]

	docType := strings.TrimRight(line1[0:2], "<")
	issuing := line1[2:5]
	surname, given := ParseName(line1[5:36])

	documentNumber := line2[0:9]
	documentNumberCheck := line2[9]
	nationality := line2[10:13]
	birth := line2[13:19]
	birthCheck := line2[19]
	sex := line2[20:21]
	expiry := line2[21:27]
	expiryCheck := line2[27]
	optional := line2[28:35]
	composite := line2[35]

	data := &MRZData{
		Raw:            strings.Join(lines, "\n"),
		Format:         format,
		DocumentType:   docType,
		IssuingCountry: issuing,
		Surname:        surname,
		GivenNames:     given,
		DocumentNumber: Clean(documentNumber),
		Nationality:    nationality,
		Sex:            Clean(sex),
		OptionalData1:  Clean(optional),
		CheckDigits: CheckDigits{
			DocumentNumber: digitValue(documentNumberCheck),
			DateOfBirth:    digitValue(birthCheck),
			ExpiryDate:     digitValue(expiryCheck),
			Composite:      digitValue(composite),
		},
	}

	var err error
	data.DateOfBirth, err = parseDate(birth)
	if err != nil {
		return nil, err
	}
	data.ExpiryDate, err = parseDate(expiry)
	if err != nil {
		return nil, err
	}

	docOK := checkMatch(documentNumber, data.CheckDigits.DocumentNumber)
	birthOK := checkMatch(birth, data.CheckDigits.DateOfBirth)
	expiryOK := checkMatch(expiry, data.CheckDigits.ExpiryDate)
	compositeSource := line2[0:10] + line2[13:20] + line2[21:35]
	compositeOK := checkMatch(compositeSource, data.CheckDigits.Composite)
	data.IsValid = docOK && birthOK && expiryOK && compositeOK

	return data, nil
}

func parseTD3(lines []string, format MRZFormat) (*MRZData, error) {
	if len(lines) != 2 || len(lines[0]) != 44 || len(lines[1]) != 44 {
		return nil, ErrInvalidFormat
	}

	line1 := lines[0]
	line2 := lines[1]

	docType := strings.TrimRight(line1[0:2], "<")
	issuing := line1[2:5]
	surname, given := ParseName(line1[5:44])

	documentNumber := line2[0:9]
	documentNumberCheck := line2[9]
	nationality := line2[10:13]
	birth := line2[13:19]
	birthCheck := line2[19]
	sex := line2[20:21]
	expiry := line2[21:27]
	expiryCheck := line2[27]
	optional := line2[28:42]
	optionalCheck := line2[42]
	composite := line2[43]

	data := &MRZData{
		Raw:            strings.Join(lines, "\n"),
		Format:         format,
		DocumentType:   docType,
		IssuingCountry: issuing,
		Surname:        surname,
		GivenNames:     given,
		DocumentNumber: Clean(documentNumber),
		Nationality:    nationality,
		Sex:            Clean(sex),
		OptionalData1:  Clean(optional),
		CheckDigits: CheckDigits{
			DocumentNumber: digitValue(documentNumberCheck),
			DateOfBirth:    digitValue(birthCheck),
			ExpiryDate:     digitValue(expiryCheck),
			OptionalData:   digitValue(optionalCheck),
			Composite:      digitValue(composite),
		},
	}

	var err error
	data.DateOfBirth, err = parseDate(birth)
	if err != nil {
		return nil, err
	}
	data.ExpiryDate, err = parseDate(expiry)
	if err != nil {
		return nil, err
	}

	docOK := checkMatch(documentNumber, data.CheckDigits.DocumentNumber)
	birthOK := checkMatch(birth, data.CheckDigits.DateOfBirth)
	expiryOK := checkMatch(expiry, data.CheckDigits.ExpiryDate)
	optionalOK := checkMatch(optional, data.CheckDigits.OptionalData)
	compositeSource := line2[0:10] + line2[13:20] + line2[21:43]
	compositeOK := checkMatch(compositeSource, data.CheckDigits.Composite)
	data.IsValid = docOK && birthOK && expiryOK && optionalOK && compositeOK

	return data, nil
}

func parseDate(value string) (time.Time, error) {
	if len(value) != 6 {
		return time.Time{}, ErrInvalidFormat
	}
	year, err := parseNumber(value[0:2])
	if err != nil {
		return time.Time{}, err
	}
	month, err := parseNumber(value[2:4])
	if err != nil {
		return time.Time{}, err
	}
	day, err := parseNumber(value[4:6])
	if err != nil {
		return time.Time{}, err
	}

	if month < 1 || month > 12 {
		return time.Time{}, ErrInvalidDate
	}

	now := time.Now().Year() % 100
	century := 1900
	if year <= now+5 {
		century = 2000
	}

	fullYear := century + year
	if day < 1 || day > daysInMonth(fullYear, month) {
		return time.Time{}, ErrInvalidDate
	}

	return time.Date(fullYear, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func parseNumber(value string) (int, error) {
	number := 0
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0, ErrInvalidFormat
		}
		number = number*10 + int(char-'0')
	}
	return number, nil
}

func daysInMonth(year int, month int) int {
	if month < 1 || month > 12 {
		return 0
	}
	nextMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC)
	return nextMonth.Day()
}
