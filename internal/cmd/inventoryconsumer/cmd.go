package inventoryconsumer

import (
	"config-manager/internal"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/util"
	"context"
	"encoding/json"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

var Command ffcli.Command = ffcli.Command{
	Name:      "inventory-consumer",
	ShortHelp: "Run the inventory kafka consumer",
	LongHelp:  "Consumes messages from the 'kafka-inventory-topic' topic and attempts to configure the identified hosts for remote configuration management.",
	Exec: func(ctx context.Context, args []string) error {
		log.Info().Str("command", "inventory-consumer").Msg("started consumer. Awaiting messages.")

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

// InventoryEvent represents a message read off the inventory.events
// topic.
type InventoryEvent struct {
	Type      string        `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Host      internal.Host `json:"host"`
}

func handler(ctx context.Context, msg kafka.Message) {
	logger := log.With().Str("module", "inventory-consumer").Int64("offset", msg.Offset).Logger()

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

	if !event.Timestamp.IsZero() && time.Since(event.Timestamp) > config.DefaultConfig.StaleEventDuration {
		logger.Info().Msg("skipping stale inventory event")
		return
	}

	// Process a message only if a host is connected via Cloud Connector
	isConnected := event.Host.SystemProfile.RHCID != ""

	if isConnected {
		switch eventType {
		case "created", "updated":
			reqID, _ := util.Kafka.GetHeader(msg, "request_id")
			logger = logger.With().Str("request_id", reqID).Str("host_id", event.Host.ID).Str("org_id", event.Host.OrgID).Logger()
			var defaultState map[string]string
			if err := json.Unmarshal([]byte(config.DefaultConfig.ServiceConfig), &defaultState); err != nil {
				logger.Error().Err(err).Msg("cannot unmarshal service config")
				return
			}

			profile, err := db.GetOrInsertCurrentProfile(event.Host.OrgID, db.NewProfile(event.Host.OrgID, event.Host.Account, defaultState))
			if err != nil {
				logger.Error().Err(err).Msg("cannot get profile from database")
				return
			}

			if !profile.OrgID.Valid {
				logger.Error().Str("account_number", db.JSONNullStringSafeValue(profile.AccountID)).Msg("profile missing org ID")
			}
		}
	}
}
