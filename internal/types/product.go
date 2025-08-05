package types

import "time"

type ProductFilter struct {
	Slug        string `json:"slug,omitempty"`
	Name        string `json:"name,omitempty"`
	Subtitle    string `json:"subtitle,omitempty"`
	Description string `json:"description,omitempty"`
	ShowAll     bool   `json:"showAll,omitempty"`
	
	CreatedAfter  time.Time `json:"createdAfter,omitempty"`
	CreatedBefore time.Time `json:"createdBefore,omitempty"`
	UpdatedAfter  time.Time `json:"updatedAfter,omitempty"`
	UpdatedBefore time.Time `json:"updatedBefore,omitempty"`
	
	Search string `json:"search,omitempty"`
	
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	
	SortBy    string `json:"sortBy,omitempty"`
	SortOrder string `json:"sortOrder,omitempty"`
	
	IncludeVariants bool `json:"includeVariants,omitempty"`
}

func DefaultProductFilter() *ProductFilter {
	return &ProductFilter{
		Limit:           20,
		Offset:          0,
		SortBy:          "createdAt",
		SortOrder:       "desc",
		IncludeVariants: true,
	}
}

func (f *ProductFilter) Validate() error {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "desc"
	}
	return nil
}

func (f *ProductFilter) IsEmpty() bool {
	var zeroTime time.Time
	
	return f.Slug == "" &&
		f.Name == "" &&
		f.Subtitle == "" &&
		f.Description == "" &&
		f.CreatedAfter.Equal(zeroTime) &&
		f.CreatedBefore.Equal(zeroTime) &&
		f.UpdatedAfter.Equal(zeroTime) &&
		f.UpdatedBefore.Equal(zeroTime) &&
		f.Search == ""
}