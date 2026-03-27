package wrappers

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/oklog/ulid/v2"
)

type ULID ulid.ULID

func (id *ULID) MarshalJSON() ([]byte, error) {
	return json.Marshal((ulid.ULID)(*id).String())
}

func (id *ULID) UnmarshalJSON(buff []byte) error {
	var tmp string
	if err := json.Unmarshal(buff, &tmp); err != nil {
		return err
	}

	if tmp == "" {
		*id = (ULID)(ulid.Zero)
		return nil
	}

	if result, err := ulid.Parse(tmp); err != nil {
		return err
	} else {
		*id = (ULID)(result)
	}
	return nil
}

func (id *ULID) Value() (driver.Value, error) {
	return (ulid.ULID)(*id).String(), nil
}

func (id *ULID) String() string {
	return (ulid.ULID)(*id).String()
}

func (id *ULID) parse(s string) error {
	if s == "" {
		*id = ULID(ulid.Zero)
		return nil
	}

	parsed, err := ulid.Parse(s)
	if err != nil {
		return err
	}

	*id = ULID(parsed)
	return nil
}

func (id *ULID) Scan(src any) error {
	switch v := src.(type) {
	case string:
		return id.parse(v)
	case []byte:
		return id.parse(string(v))
	case nil:
		*id = ULID(ulid.Zero)
		return nil
	default:
		return fmt.Errorf("wrappers.ULID.Scan: unsupported type %T", src)
	}
}
