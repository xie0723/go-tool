package ctime

import (
	"errors"
	"time"
)

type Time time.Time

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	var err error
	var tt time.Time
	if tt, err = time.Parse(`"2006-01-02T15:04:05Z0700"`, string(data)); err == nil {
		if (tt != time.Time{}) {
			*t = Time(tt)
			return nil
		}
	}
	if tt, err = time.Parse(`"2006-01-02 15:04:05"`, string(data)); err == nil {
		if (tt != time.Time{}) {
			*t = Time(tt)
			return nil
		}
	}
	return err
}

// MarshalJSON implements the json.Marshaler interface.
// The time is a quoted string in RFC 3339 format, with sub-second precision added if present.
func (tt Time) MarshalJSON() ([]byte, error) {
	t := time.Time(tt)
	if y := t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(time.RFC3339Nano)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, time.RFC3339Nano)
	b = append(b, '"')
	return b, nil
}
