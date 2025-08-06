package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// CustomerInfo implements the driver.Valuer and sql.Scanner interfaces
// for storing/retrieving from JSONB columns

// Value implements driver.Valuer for CustomerInfo
func (c CustomerInfo) Value() (driver.Value, error) {
	if c == (CustomerInfo{}) {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements sql.Scanner for CustomerInfo
func (c *CustomerInfo) Scan(value interface{}) error {
	if value == nil {
		*c = CustomerInfo{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprintf("cannot scan %T into CustomerInfo", value))
	}

	return json.Unmarshal(bytes, c)
}

// DeliveryInfo implements the driver.Valuer and sql.Scanner interfaces

// Value implements driver.Valuer for DeliveryInfo
func (d DeliveryInfo) Value() (driver.Value, error) {
	if d == (DeliveryInfo{}) {
		return nil, nil
	}
	return json.Marshal(d)
}

// Scan implements sql.Scanner for DeliveryInfo
func (d *DeliveryInfo) Scan(value interface{}) error {
	if value == nil {
		*d = DeliveryInfo{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprintf("cannot scan %T into DeliveryInfo", value))
	}

	return json.Unmarshal(bytes, d)
}

// Money implements the driver.Valuer and sql.Scanner interfaces

// Value implements driver.Valuer for Money
func (m Money) Value() (driver.Value, error) {
	if m == (Money{}) {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner for Money
func (m *Money) Scan(value interface{}) error {
	if value == nil {
		*m = Money{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprintf("cannot scan %T into Money", value))
	}

	return json.Unmarshal(bytes, m)
}

// MoneyMap and CaviarDetails driver methods are already implemented in product_driver.go