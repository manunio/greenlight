package data

import "github.com/manunio/greenlight/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_00_000, "page", "must be greater than zero")
	v.Check(f.Page <= 10_00_000, "page", "must be a maximum of 10 million")
	v.Check(f.Page > 0, "page_size", "must be greater than zero")
	v.Check(f.Page > 0, "page_size", "must be a maximum of 100")

	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
