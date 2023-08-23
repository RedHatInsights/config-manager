package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID           uuid.UUID       `json:"id" db:"profile_id"`
	Name         *JSONNullString `json:"name,omitempty" db:"name"`
	Label        *JSONNullString `json:"label,omitempty" db:"label"`
	AccountID    *JSONNullString `json:"account_id,omitempty" db:"account_id"`
	OrgID        *JSONNullString `json:"org_id,omitempty" db:"org_id"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	Active       bool            `json:"active" db:"active"`
	Creator      *JSONNullString `json:"creator,omitempty" db:"creator"`
	Insights     bool            `json:"insights" db:"insights"`
	Remediations bool            `json:"remediations" db:"remediations"`
	Compliance   bool            `json:"compliance" db:"compliance"`
}

// NewProfile creates a new, default Profile.
func NewProfile(orgID string, accountID string, state map[string]string) *Profile {
	profile := Profile{
		ID:        uuid.New(),
		Label:     &JSONNullString{NullString: sql.NullString{Valid: true, String: accountID + "-" + uuid.New().String()}},
		AccountID: &JSONNullString{NullString: sql.NullString{Valid: accountID != "", String: accountID}},
		OrgID:     &JSONNullString{NullString: sql.NullString{Valid: orgID != "", String: orgID}},
		Creator:   &JSONNullString{NullString: sql.NullString{Valid: true, String: "redhat"}},
		CreatedAt: time.Now(),
		Active:    true,
	}
	profile.SetStateConfig(state)

	return &profile
}

// CopyProfile creates a new Profile, copying values from the given Profile
// where appropriate.
func CopyProfile(from Profile) Profile {
	return Profile{
		ID:   uuid.New(),
		Name: from.Name,
		Label: func() *JSONNullString {
			var val string
			if from.AccountID != nil && from.AccountID.Valid {
				val = from.AccountID.String + "-"
			}
			val = val + uuid.New().String()
			return &JSONNullString{
				NullString: sql.NullString{
					Valid:  true,
					String: val,
				},
			}
		}(),
		AccountID:    from.AccountID,
		OrgID:        from.OrgID,
		CreatedAt:    time.Now(),
		Active:       from.Active,
		Creator:      from.Creator,
		Insights:     from.Insights,
		Remediations: from.Remediations,
		Compliance:   from.Compliance,
	}
}

func (p Profile) Equal(q Profile) bool {
	return p.Active == q.Active && p.Insights == q.Insights && p.Compliance == q.Compliance && p.Remediations == q.Remediations
}

// StateConfig formats the profile's state values as a "state map".
func (p Profile) StateConfig() map[string]string {
	state := make(map[string]string)

	if p.Insights {
		state["insights"] = "enabled"
	} else {
		state["insights"] = "disabled"
	}

	if p.Remediations {
		state["remediations"] = "enabled"
	} else {
		state["remediations"] = "disabled"
	}

	if p.Compliance {
		state["compliance_openscap"] = "enabled"
	} else {
		state["compliance_openscap"] = "disabled"
	}

	return state
}

// SetStateConfig parses a "state map" and updates the profile's state values
// accordingly.
func (p *Profile) SetStateConfig(state map[string]string) {
	if v, has := state["insights"]; has {
		p.Insights = v == "enabled"
	}

	if v, has := state["remediations"]; has {
		p.Remediations = v == "enabled"
	}

	if v, has := state["compliance_openscap"]; has {
		p.Compliance = v == "enabled"
	}
}

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

// JSONNullString represents a string that may be null simultaneously in a SQL
// data field and a JSON value. JSONNullString implements the json.Marshaler and
// json.Unmarshaler interfaces so it can be marshalled and unmarshalled to and
// from a JSON value.
type JSONNullString struct {
	sql.NullString
}

// JSONNullStringSafeValue returns the string value of n if it is a valid value,
// otherwise it returns "".
func JSONNullStringSafeValue(n *JSONNullString) string {
	if n != nil && n.Valid {
		return n.String
	}
	return ""
}

func (n JSONNullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.String)
	}
	return []byte(`""`), nil
}

func (n *JSONNullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.String = ""
		return nil
	}
	err := json.Unmarshal(data, &n.String)
	if err != nil {
		return err
	}
	n.Valid = err == nil

	return nil
}
