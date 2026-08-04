// Package locale defines the closed set of locales supported by the MVP.
package locale

type Locale string

const (
	EnglishUS    Locale = "EN_US"
	EnglishCA    Locale = "EN_CA"
	FrenchCA     Locale = "FR_CA"
	PortugueseBR Locale = "PT_BR"
)

func (l Locale) Valid() bool {
	switch l {
	case EnglishUS, EnglishCA, FrenchCA, PortugueseBR:
		return true
	default:
		return false
	}
}
