package db

import "database/sql"

// JSONNullBool represents a bool that may be null simultaneously in a SQL
// data field and a JSON value. JSONNullBool implements the json.Marshaler
// and json.Unmarshaler interfaces so it can be marshalled and unmarshalled to
// and from a JSON value.
type JSONNullBool struct {
	sql.NullBool
}

func (n JSONNullBool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		if n.Bool {
			return []byte(`true`), nil
		}
		return []byte(`false`), nil
	}
	return []byte(`null`), nil
}

func (n *JSONNullBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `true`:
		n.Valid = true
		n.Bool = true
	case `false`:
		n.Valid = true
		n.Bool = false
	default:
		n.Valid = false
		n.Bool = false
	}

	return nil
}
