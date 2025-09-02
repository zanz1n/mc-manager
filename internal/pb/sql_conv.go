package pb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

var _ driver.Valuer = (*InstanceConfig)(nil)
var _ sql.Scanner = (*InstanceConfig)(nil)

// Scan implements sql.Scanner.
func (x *InstanceConfig) Scan(src any) error {
	switch src := src.(type) {
	case string:
		if err := json.Unmarshal([]byte(src), x); err != nil {
			return scanErrInstanceConfig(src, err)
		}

	case []byte:
		if err := json.Unmarshal(src, x); err != nil {
			return scanErrInstanceConfig(src, err)
		}

	default:
		return scanErrInstanceConfig(src, nil)
	}
	return nil
}

// Value implements driver.Valuer.
func (x *InstanceConfig) Value() (driver.Value, error) {
	return json.Marshal(x)
}

var _ sql.Scanner = (*InstanceLimits)(nil)
var _ driver.Valuer = (*InstanceLimits)(nil)

// Value implements driver.Valuer.
func (x *InstanceLimits) Value() (driver.Value, error) {
	return json.Marshal(x)
}

// Scan implements sql.Scanner.
func (x *InstanceLimits) Scan(src any) error {
	switch src := src.(type) {
	case string:
		if err := json.Unmarshal([]byte(src), x); err != nil {
			return scanErrInstanceLimits(src, err)
		}

	case []byte:
		if err := json.Unmarshal(src, x); err != nil {
			return scanErrInstanceLimits(src, err)
		}

	default:
		return scanErrInstanceLimits(src, nil)
	}
	return nil
}

func scanErrInstanceConfig(src any, err error) error {
	if err != nil {
		return fmt.Errorf(
			"Scan: unable to scan type %T into InstanceConfig",
			src,
		)
	} else {
		return fmt.Errorf(
			"Scan: unable to scan type %T into InstanceConfig: %w",
			src,
			err,
		)
	}
}

func scanErrInstanceLimits(src any, err error) error {
	if err != nil {
		return fmt.Errorf(
			"Scan: unable to scan type %T into InstanceLimits",
			src,
		)
	} else {
		return fmt.Errorf(
			"Scan: unable to scan type %T into InstanceLimits: %w",
			src,
			err,
		)
	}
}
