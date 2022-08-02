package inventoryconsumer

import (
	"config-manager/infrastructure"
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
	container := infrastructure.Container{}
	logger := log.With().Str("module", "inventory-consumer").Logger()

	eventType, err := util.Kafka.GetHeader(msg, "event_type")
	if err != nil {
		logger.Error().Err(err).Msg("error getting event_type")
		return
	}
	logger.Trace().Msgf("event_type = %v", eventType)

	if eventType == "created" || eventType == "updated" {
		value := &InventoryEvent{}

		if err := json.Unmarshal(msg.Value, &value); err != nil {
			logger.Error().Err(err).Msg("couldn't unmarshal inventory event")
			return
		}

		if value.Host.Reporter == "cloud-connector" {
			if eventType == "created" {
				logger.Info().Msg("new host detected; setting up for playbook execution")
				messageID, err := container.CMService().SetupHost(ctx, value.Host)
				if err != nil {
					logger.Error().Err(err).Msgf("error setting up host: %v", value.Host)
					return
				}
				logger.Info().Msgf("Cloud-connector setup host message id: %v", messageID)
			}

			var defaultState map[string]string
			if err := json.Unmarshal([]byte(config.DefaultConfig.ServiceConfig), &defaultState); err != nil {
				log.Printf("cannot unmarshal data: %v", err)
				return
			}
			profile, err := db.GetOrInsertCurrentProfile(value.Host.OrgID, db.NewProfile(value.Host.OrgID, value.Host.Account, defaultState))
			if err != nil {
				logger.Error().Err(err).Msgf("Error retrieving state for account: %v", value.Host.Account)
				return
			}

			if !profile.OrgID.Valid {
				logger.Debug().Msg("profile missing org ID")
				if config.DefaultConfig.TenantTranslatorHost != "" {
					translator := tenantid.NewTranslator(config.DefaultConfig.TenantTranslatorHost)
					orgID, err := translator.EANToOrgID(ctx, profile.AccountID.String)
					if err != nil {
						logger.Error().Err(err).Msg("unable to translate EAN to orgID")
						return
					}
					logger.Debug().Str("org_id", orgID).Str("account_number", profile.AccountID.String).Msg("translated EAN to orgID")
					profile.OrgID.Valid = orgID != ""
					profile.OrgID.String = orgID

					if err := db.InsertProfile(*profile); err != nil {
						log.Error().Err(err).Msg("unable to insert profile")
						return
					}
					logger.Debug().Msg("inserted new profile")
				}
			}

			reqID, err := util.Kafka.GetHeader(msg, "request_id")
			if err != nil {
				logger.Error().Err(err).Msg("Error getting request_id header")
				k := requestIDkey("request_id")
				reqID = uuid.New().String()
				logger.Info().Msgf("Creating new request_id and adding to context: %v", reqID)
				ctx = context.WithValue(ctx, k, reqID)
			}
			logger = logger.With().Str("request_id", reqID).Logger()

			logger.Info().Msgf("Cloud-connector inventory event request_id: %s, data: %+v", reqID, value)

			if value.Host.SystemProfile.RHCState != profile.ID.String() {
				logger.Info().Msgf("rhc_state_id %s for client %s does not match current state id %s for account %s. Updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, profile.ID.String(), profile.AccountID.String)
				client := []internal.Host{value.Host}
				responses, err := container.CMService().ApplyState(ctx, *profile, client)
				if err != nil {
					logger.Error().Err(err).Msg("error applying state")
				}
				logger.Info().Msgf("Message sent to the dispatcher. Results: %v", responses)
			} else {
				logger.Info().Msgf("rhc_state_id %s for client %s is up to date for account %s. Not updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, profile.AccountID.String)
			}
		}
	}
}
