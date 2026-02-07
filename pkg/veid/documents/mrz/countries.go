package mrz

import "strings"

var countryNames = map[string]string{
	"AUS": "Australia",
	"ARE": "United Arab Emirates",
	"AUT": "Austria",
	"BEL": "Belgium",
	"BGR": "Bulgaria",
	"BRA": "Brazil",
	"CAN": "Canada",
	"CHE": "Switzerland",
	"CHN": "China",
	"CYP": "Cyprus",
	"CZE": "Czechia",
	"DEU": "Germany",
	"DNK": "Denmark",
	"ESP": "Spain",
	"EST": "Estonia",
	"FIN": "Finland",
	"FRA": "France",
	"GBR": "United Kingdom",
	"GRC": "Greece",
	"HRV": "Croatia",
	"HUN": "Hungary",
	"IND": "India",
	"IRL": "Ireland",
	"ITA": "Italy",
	"JPN": "Japan",
	"KOR": "South Korea",
	"LIE": "Liechtenstein",
	"LTU": "Lithuania",
	"LUX": "Luxembourg",
	"LVA": "Latvia",
	"MLT": "Malta",
	"NLD": "Netherlands",
	"POL": "Poland",
	"PRT": "Portugal",
	"ROU": "Romania",
	"SVK": "Slovakia",
	"SVN": "Slovenia",
	"SWE": "Sweden",
	"USA": "United States",
	"UTO": "Utopia",
}

var alpha2ToAlpha3 = map[string]string{
	"AU": "AUS",
	"BR": "BRA",
	"CA": "CAN",
	"CN": "CHN",
	"DE": "DEU",
	"ES": "ESP",
	"FR": "FRA",
	"GB": "GBR",
	"GR": "GRC",
	"IE": "IRL",
	"IN": "IND",
	"IT": "ITA",
	"JP": "JPN",
	"KR": "KOR",
	"NL": "NLD",
	"PT": "PRT",
	"SE": "SWE",
	"US": "USA",
}

// NormalizeCountryCode ensures uppercase ISO codes and maps alpha-2 to alpha-3 when possible.
func NormalizeCountryCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ""
	}
	if len(code) == 2 {
		if alpha3, ok := alpha2ToAlpha3[code]; ok {
			return alpha3
		}
	}
	return code
}

// CountryName returns a human-readable country name for a 3-letter code.
func CountryName(code string) (string, bool) {
	code = NormalizeCountryCode(code)
	name, ok := countryNames[code]
	return name, ok
}
