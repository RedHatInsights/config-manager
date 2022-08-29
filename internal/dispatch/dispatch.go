package dispatch

import (
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/infrastructure/persistence/inventory"
	"config-manager/internal"
	"config-manager/internal/db"
	"context"

	"github.com/rs/zerolog/log"
)

var profiles chan db.Profile

func init() {
	profiles = make(chan db.Profile)
	go func() {
		for p := range profiles {
			hosts, err := inventory.NewInventoryClient().GetAllInventoryClients(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("cannot get hosts from inventory")
				continue
			}

			internal.ApplyProfile(context.Background(), &p, hosts, func(resp []dispatcher.RunCreated) {
				log.Info().Str("profile_id", p.ID.String()).Msg("applied profile")
			})
		}
	}()
}

// Dispatch queues the profile on a channel to be dispatched to connected hosts.
func Dispatch(p db.Profile) {
	go func() {
		profiles <- p
	}()
}
