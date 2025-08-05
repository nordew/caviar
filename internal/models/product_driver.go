package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

func (mm *MoneyMap) Scan(value any) error {
    if value == nil {
        *mm = make(MoneyMap)
        return nil
    }
    
    var bytes []byte
    switch v := value.(type) {
    case []byte:
        bytes = v
    case string:
        bytes = []byte(v)
    default:
        return fmt.Errorf("cannot scan %T into MoneyMap", value)
    }
    
    return json.Unmarshal(bytes, mm)
}

func (mm MoneyMap) Value() (driver.Value, error) {
    if mm == nil {
        return "{}", nil
    }
    return json.Marshal(mm)
}

func (cd *CaviarDetails) Scan(value any) error {
    if value == nil {
        return nil
    }
    
    var bytes []byte
    switch v := value.(type) {
    case []byte:
        bytes = v
    case string:
        bytes = []byte(v)
    default:
        return fmt.Errorf("cannot scan %T into CaviarDetails", value)
    }
    
    return json.Unmarshal(bytes, cd)
}

func (cd CaviarDetails) Value() (driver.Value, error) {
    return json.Marshal(cd)
}