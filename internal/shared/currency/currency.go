package currency

type Code string

const (
	CAD Code = "CAD"
	USD Code = "USD"
	BRL Code = "BRL"
)

func (c Code) Valid() bool {
	switch c {
	case CAD, USD, BRL:
		return true
	default:
		return false
	}
}
