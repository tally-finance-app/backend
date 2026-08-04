package pagination

type PageSize int

const (
	PageSize10  PageSize = 10
	PageSize25  PageSize = 25
	PageSize50  PageSize = 50
	PageSize100 PageSize = 100
	PageSize200 PageSize = 200

	DefaultPageSize = PageSize50
)

func (p PageSize) Valid() bool {
	switch p {
	case PageSize10, PageSize25, PageSize50, PageSize100, PageSize200:
		return true
	default:
		return false
	}
}
