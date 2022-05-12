package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID           uuid.UUID      `db:"profile_id"`
	Name         sql.NullString `db:"name"`
	Label        sql.NullString `db:"label"`
	AccountID    sql.NullString `db:"account_id"`
	OrgID        sql.NullString `db:"org_id"`
	CreatedAt    time.Time      `db:"created_at"`
	Active       bool           `db:"active"`
	Creator      sql.NullString `db:"creator"`
	Insights     bool           `db:"insights"`
	Remediations bool           `db:"remediations"`
	Compliance   bool           `db:"compliance"`
}

// NewProfile creates a new, default Profile.
func NewProfile(accountID string, state map[string]string) *Profile {
	profile := Profile{
		ID:        uuid.New(),
		Label:     sql.NullString{Valid: true, String: accountID + "-" + uuid.New().String()},
		AccountID: sql.NullString{Valid: true, String: accountID},
		Creator:   sql.NullString{Valid: true, String: "redhat"},
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
		ID:           uuid.New(),
		Name:         from.Name,
		Label:        sql.NullString{Valid: true, String: from.AccountID.String + "-" + uuid.New().String()},
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
