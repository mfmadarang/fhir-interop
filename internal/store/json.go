package store

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSON []byte

// tells GORM's migrator to use the jsonb column type on Postgres
func (JSON) GormDataType() string {
	return "jsonb"
}

// implements driver.Valuer so GORM can write
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// implements sql.Scanner so GORM can read
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
		return nil
	case string:
		*j = JSON(v)
		return nil
	default:
		return fmt.Errorf("unsupported type for JSON scan: %T", value)
	}
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSON: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

func toJSON(v interface{}) (JSON, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return JSON(b), nil
}