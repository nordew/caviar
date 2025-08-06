package types

import "time"

type OrderFilter struct {
	Status        string
	CustomerPhone string
	Country       string
	CreatedFrom   time.Time
	CreatedTo     time.Time
	Limit         int
	Offset        int
}