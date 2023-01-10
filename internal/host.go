package internal

import (
	"config-manager/infrastructure/persistence/cloudconnector"
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/util"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// Host represents a system record from the Inventory application.
type Host struct {
	ID                    string                 `json:"id"`
	Account               string                 `json:"account"`
	OrgID                 string                 `json:"org_id"`
	DisplayName           string                 `json:"display_name"`
	Reporter              string                 `json:"reporter"`
	PerReporterStaleness  map[string]interface{} `json:"per_reporter_staleness"`
	SubscriptionManagerID string                 `json:"subscription_manager_id"`
	SystemProfile         struct {
		RHCID    string `json:"rhc_client_id"`
		RHCState string `json:"rhc_config_state"`
	} `json:"system_profile"`
}

// ApplyProfile applies the current profile to the specified hosts.
func ApplyProfile(ctx context.Context, profile *db.Profile, hosts []Host, fn func(resp []dispatcher.RunCreated)) {
	logger := log.With().Logger()
	if profile.AccountID != nil && profile.AccountID.Valid {
		logger = logger.With().Str("account_id", profile.AccountID.String).Logger()
	}
	if profile.OrgID != nil && profile.OrgID.Valid {
		logger = logger.With().Str("org_id", profile.OrgID.String).Logger()
	}

	if !profile.Active {
		logger.Info().Interface("profile", profile).Msg("skipping application of inactive profile")
		return
	}

	logger.Debug().Int("num_hosts", len(hosts)).Msg("applying profile for hosts")

	runs := make([]dispatcher.RunInput, 0, len(hosts))
	for _, host := range hosts {
		if _, has := host.PerReporterStaleness["cloud-connector"]; !has {
			logger.Warn().Str("client_id", host.SystemProfile.RHCID).Msg("detected host without cloud-connector as a reporter")
			continue
		}
		logger.Debug().Str("client_id", host.SystemProfile.RHCID).Msg("creating run for host")
		run := dispatcher.RunInput{
			Recipient: host.SystemProfile.RHCID,
			Account:   profile.AccountID.String,
			Url:       config.DefaultConfig.PlaybookHost.String() + fmt.Sprintf(config.DefaultConfig.PlaybookPath, profile.ID),
			Labels: &dispatcher.RunInput_Labels{
				AdditionalProperties: map[string]string{
					"state_id": profile.ID.String(),
					"id":       host.ID,
				},
			},
		}
		runs = append(runs, run)
	}

	if err := util.Batch.All(len(runs), config.DefaultConfig.DispatcherBatchSize, func(start, end int) error {
		go func() {
			log.Debug().Int("start", start).Int("end+1", end+1).Msg("batching runs")

			resp, err := dispatcher.NewDispatcherClient().Dispatch(ctx, runs[start:end+1])
			if err != nil {
				logger.Error().Err(err).Msg("cannot dispatch to playbook-dispatcher")
				return
			}
			logger.Trace().Interface("runs_created", resp).Msg("dispatched work to playbook-dispatcher")
			fn(resp)
		}()
		return nil
	}); err != nil {
		logger.Error().Err(err).Msg("cannot batch work")
	}
}

// SetupHost sends a message to the host through cloud-connector to install the
// rhc-worker-playbook RPM, enabling the host to receive and execute playbooks
// sent through playbook-dispatcher.
func SetupHost(ctx context.Context, host Host) (string, error) {
	logger := log.With().Str("account_id", host.Account).Str("subscription_manager_id", host.SubscriptionManagerID).Logger()

	if host.Account == "" {
		return "", fmt.Errorf("cannot setup host: missing value for 'account' field")
	}

	if host.SubscriptionManagerID == "" {
		return "", fmt.Errorf("cannot setup host: missing value for 'subscription_manager_id' field")
	}

	client, err := cloudconnector.NewCloudConnectorClient()
	if err != nil {
		logger.Error().Err(err).Msg("cannot get cloud-connector client")
		return "", err
	}

	var status string
	var dispatchers map[string]interface{}
	started := time.Now()
	for {
		if time.Now().After(started.Add(180 * time.Second)) {
			return "", fmt.Errorf("cannot get connected status after %v, aborting", time.Since(started))
		}

		status, dispatchers, err = client.GetConnectionStatus(ctx, host.Account, host.SubscriptionManagerID)
		if err != nil {
			logger.Error().Err(err).Msg("cannot get connection status from cloud-connector")
			return "", err
		}
		logger.Debug().Str("status", status).Interface("dispatchers", dispatchers).Msg("connection status from cloud-connector")
		if status == "connected" {
			break
		}
		time.Sleep(30 * time.Second)
	}

	if status != "connected" {
		err := fmt.Errorf("host not connected")
		logger.Error().Str("status", status).Err(err).Msg("cannot setup host")
		return "", fmt.Errorf("cannot setup host: %w", err)
	}

	if _, has := dispatchers["package-manager"]; !has {
		return "", fmt.Errorf("host %v missing required directive 'package-manager'", host.SubscriptionManagerID)
	}

	if _, has := dispatchers["rhc-worker-playbook"]; has {
		return "", nil
	}

	payload := struct {
		Command string `json:"command"`
		Name    string `json:"name"`
	}{
		Command: "install",
		Name:    "rhc-worker-playbook",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		logger.Error().Err(err).Msg("cannot marshal payload")
		return "", fmt.Errorf("cannot marshal payload: %v", err)
	}

	messageID, err := client.SendMessage(ctx, host.Account, "package-manager", data, nil, host.SubscriptionManagerID)
	if err != nil {
		logger.Error().Err(err).Msg("cannot send message to host")
		return "", err
	}
	logger.Debug().Str("directive", "package-manager").Interface("payload", payload).Msg("sent message to host")

	started = time.Now()
	for {
		if time.Now().After(started.Add(180 * time.Second)) {
			return "", fmt.Errorf("unable to detect rhc-worker-playbook after %v, aborting", time.Since(started))
		}
		status, dispatchers, err := client.GetConnectionStatus(ctx, host.Account, host.SubscriptionManagerID)
		if err != nil {
			logger.Error().Err(err).Msg("cannot get connection status from cloud-connector")
			return "", err
		}
		logger.Debug().Str("status", status).Interface("dispatchers", dispatchers).Msg("connection status from cloud-connector")

		if status == "disconnected" {
			return messageID, fmt.Errorf("host disconnected while waiting for connection status")
		}
		if _, has := dispatchers["rhc-worker-playbook"]; has {
			logger.Debug().Interface("dispatchers", dispatchers).Msg("found rhc-worker-playbook dispatcher")
			break
		}
		time.Sleep(30 * time.Second)
	}

	return messageID, nil
}
