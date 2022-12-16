package inventoryconsumer

import (
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/internal"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/util"
	"context"
	"encoding/json"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"
	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

var Command ffcli.Command = ffcli.Command{
	Name:      "inventory-consumer",
	ShortHelp: "Run the inventory kafka consumer",
	LongHelp:  "Consumes messages from the 'kafka-inventory-topic' topic and attempts to configure the identified hosts for remote configuration management.",
	Exec: func(ctx context.Context, args []string) error {
		log.Info().Str("command", "inventory-consumer").Msg("starting command")

		reader := util.Kafka.NewReader(config.DefaultConfig.KafkaInventoryTopic)

		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Error().Err(err).Msg("unable to read message")
				continue
			}
			go handler(ctx, m)
		}
	},
}

type requestIDkey string

// InventoryEvent represents a message read off the inventory.events
// topic.
type InventoryEvent struct {
	Type string        `json:"type"`
	Host internal.Host `json:"host"`
}

func handler(ctx context.Context, msg kafka.Message) {
	logger := log.With().Str("module", "inventory-consumer").Logger()

	eventType, err := util.Kafka.GetHeader(msg, "event_type")
	if err != nil {
		logger.Error().Err(err).Msg("error getting event_type")
		return
	}
	logger = logger.With().Str("event_type", eventType).Logger()

	event := &InventoryEvent{}

	if err := json.Unmarshal(msg.Value, event); err != nil {
		logger.Error().Err(err).Msg("cannot unmarshal inventory event")
		return
	}

	if event.Host.Reporter != "cloud-connector" {
		logger.Debug().Str("reporter", event.Host.Reporter).Msg("ignoring host")
		return
	}

	switch eventType {
	case "created":
		logger.Info().Msg("setting up new host for remote host configuration")
		messageID, err := internal.SetupHost(ctx, event.Host)
		if err != nil {
			logger.Error().Err(err).Interface("host", event.Host).Msg("cannot set up up host")
			return
		}
		logger.Info().Str("message_id", messageID).Msg("setup message sent to host")
		fallthrough
	case "updated":
		var defaultState map[string]string
		if err := json.Unmarshal([]byte(config.DefaultConfig.ServiceConfig), &defaultState); err != nil {
			logger.Error().Err(err).Msg("cannot unmarshal service config")
			return
		}

		profile, err := db.GetOrInsertCurrentProfile(event.Host.OrgID, db.NewProfile(event.Host.OrgID, event.Host.Account, defaultState))
		if err != nil {
			logger.Error().Err(err).Str("org_id", event.Host.OrgID).Msg("cannot get profile from database")
			return
		}

		if !profile.OrgID.Valid {
			logger.Debug().Msg("profile missing org ID")
			if config.DefaultConfig.TenantTranslatorHost != "" {
				translator := tenantid.NewTranslator(config.DefaultConfig.TenantTranslatorHost)
				orgID, err := translator.EANToOrgID(ctx, profile.AccountID.String)
				if err != nil {
					logger.Error().Err(err).Msg("cannot translate EAN to orgID")
					return
				}
				logger.Debug().Str("org_id", orgID).Str("account_number", profile.AccountID.String).Msg("translated EAN to orgID")
				profile.OrgID.Valid = orgID != ""
				profile.OrgID.String = orgID

				if err := db.InsertProfile(*profile); err != nil {
					logger.Error().Err(err).Msg("cannot insert profile")
					return
				}
				logger.Debug().Msg("inserted new profile")
			}
		}

		reqID, err := util.Kafka.GetHeader(msg, "request_id")
		if err != nil {
			logger.Error().Err(err).Msg("cannot get request_id header")
			reqID = uuid.New().String()
			logger.Debug().Str("request_id", reqID).Msg("created request ID")
			ctx = context.WithValue(ctx, requestIDkey("request_id"), reqID)
		}
		logger = logger.With().Str("request_id", reqID).Logger()

		if event.Host.SystemProfile.RHCState != profile.ID.String() {
			logger.Info().Str("host.system_profile.rhc_config_state", event.Host.SystemProfile.RHCState).Str("profile.id", profile.ID.String()).Interface("host", event.Host).Msg("updating state configuration for host")
			host := []internal.Host{event.Host}
			internal.ApplyProfile(ctx, profile, host, func(responses []dispatcher.RunCreated) {
				logger.Info().Interface("responses", responses).Msg("received response from playbook-dispatcher")
			})
		} else {
			logger.Info().Str("host.system_profile.rhc_config_state", event.Host.SystemProfile.RHCState).Str("profile.id", profile.ID.String()).Interface("host", event.Host).Msg("host state matches profile ID")
		}
	}
}
