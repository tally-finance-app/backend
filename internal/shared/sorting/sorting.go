package sorting

type Direction string

const (
	Ascending  Direction = "asc"
	Descending Direction = "desc"
)

func (d Direction) Valid() bool {
	switch d {
	case Ascending, Descending:
		return true
	default:
		return false
	}
}
