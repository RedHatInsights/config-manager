package dispatch

import (
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/infrastructure/persistence/inventory"
	"config-manager/internal"
	"config-manager/internal/db"
	"context"

	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/rs/zerolog/log"
)

type ProfileWithIdentity struct {
	Profile  db.Profile
	Identity identity.XRHID
}

var profileIdentityChan chan ProfileWithIdentity

func init() {
	profileIdentityChan = make(chan ProfileWithIdentity)
	go func() {
		for profileIdentity := range profileIdentityChan {
			ctx := context.WithValue(context.Background(), identity.Key, profileIdentity.Identity)
			hosts, err := inventory.NewInventoryClient().GetAllInventoryClients(ctx)
			if err != nil {
				log.Error().Err(err).Msg("cannot get hosts from inventory")
				continue
			}

			internal.ApplyProfile(ctx, &profileIdentity.Profile, hosts, func(resp []dispatcher.RunCreated) {
				log.Info().Str("profile_id", profileIdentity.Profile.ID.String()).Msg("applied profile")
			})
		}
	}()
}

// Dispatch queues the profile with identity on a channel to be dispatched to connected hosts.
func Dispatch(profile db.Profile, identity identity.XRHID) {
	go func() {
		profileIdentityChan <- ProfileWithIdentity{Profile: profile, Identity: identity}
	}()
}
