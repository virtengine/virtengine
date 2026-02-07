package adapters

import "github.com/virtengine/virtengine/pkg/veid/documents"

var DefaultCountries = []documents.CountryCode{
	// EU (selected ISO 3166-1 alpha-3)
	"AUT", "BEL", "BGR", "HRV", "CYP", "CZE", "DNK", "EST", "FIN", "FRA",
	"DEU", "GRC", "HUN", "IRL", "ITA", "LVA", "LTU", "LUX", "MLT", "NLD",
	"POL", "PRT", "ROU", "SVK", "SVN", "ESP", "SWE",
	// UK + US + CA + AU + JP + KR + IN + UAE + BR
	"GBR", "USA", "CAN", "AUS", "JPN", "KOR", "IND", "ARE", "BRA",
}

func DefaultRegistry() *documents.Registry {
	return documents.NewRegistry(NewMRZAdapter(DefaultCountries))
}
