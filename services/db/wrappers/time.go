package wrappers

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Time time.Time

func (t *Time) MarshalJSON() ([]byte, error) {
	return (time.Time)(*t).MarshalJSON()
}

func (t *Time) UnmarshalJSON(buff []byte) error {
	var tmp string
	if err := json.Unmarshal(buff, &tmp); err != nil {
		return err
	}

	if result, err := time.Parse(time.RFC3339Nano, tmp); err != nil {
		return err
	} else {
		*t = (Time)(result)
	}
	return nil
}

func (t Time) String() string {
	return time.Time(t).Format(time.RFC3339Nano)
}

func (t *Time) Value() (driver.Value, error) {
	return time.Time(*t), nil
}

func (t *Time) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		*t = Time(v)
		return nil
	case string:
		parsed, err := parseTime(v)
		if err != nil {
			return err
		}
		*t = Time(parsed)
		return nil
	case []byte:
		parsed, err := parseTime(string(v))
		if err != nil {
			return err
		}
		*t = Time(parsed)
		return nil
	case nil:
		*t = Time(time.Time{})
		return nil
	default:
		return fmt.Errorf("wrappers.Time.Scan: unsupported type %T", src)
	}
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}
