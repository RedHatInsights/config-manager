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
	ID                   string                 `json:"id"`
	Account              string                 `json:"account"`
	OrgID                string                 `json:"org_id"`
	DisplayName          string                 `json:"display_name"`
	Reporter             string                 `json:"reporter"`
	PerReporterStaleness map[string]interface{} `json:"per_reporter_staleness"`
	SystemProfile        struct {
		RHCID    string `json:"rhc_client_id"`
		RHCState string `json:"rhc_config_state"`
	} `json:"system_profile"`
}

// ApplyProfile applies the current profile to the specified hosts.
func ApplyProfile(ctx context.Context, profile *db.Profile, hosts []Host, fn func(resp []dispatcher.RunCreated)) {
	logger := log.With().Str("account_id", profile.AccountID.String).Str("org_id", profile.OrgID.String).Logger()

	if !profile.Active {
		logger.Info().Interface("profile", profile).Msg("skipping application of inactive profile")
		return
	}

	logger.Debug().Int("num_hosts", len(hosts)).Msg("applying profile for hosts")

	runs := make([]dispatcher.RunInput, 0, len(hosts))
	for _, host := range hosts {
		if _, has := host.PerReporterStaleness["cloud-connector"]; !has {
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
	logger := log.With().Str("account_id", host.Account).Str("client_id", host.SystemProfile.RHCID).Logger()

	client, err := cloudconnector.NewCloudConnectorClient()
	if err != nil {
		logger.Error().Err(err).Msg("cannot get cloud-connector client")
		return "", err
	}
	status, dispatchers, err := client.GetConnectionStatus(ctx, host.Account, host.SystemProfile.RHCID)
	if err != nil {
		logger.Error().Err(err).Msg("cannot get connection status from cloud-connector")
		return "", err
	}
	logger.Debug().Str("status", status).Interface("dispatchers", dispatchers).Msg("connection status from cloud-connector")

	if status != "connected" {
		return "", fmt.Errorf("cannot set up host: host connection status = %v", status)
	}

	if _, has := dispatchers["package-manager"]; !has {
		return "", fmt.Errorf("host %v missing required directive 'package-manager'", host.SystemProfile.RHCID)
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

	messageID, err := client.SendMessage(ctx, host.Account, "package-manager", data, nil, host.SystemProfile.RHCID)
	if err != nil {
		logger.Error().Err(err).Msg("cannot send message to host")
		return "", err
	}

	started := time.Now()
	for {
		if time.Now().After(started.Add(180 * time.Second)) {
			return "", fmt.Errorf("unable to detect rhc-worker-playbook after %v, aborting", time.Since(started))
		}
		status, dispatchers, err := client.GetConnectionStatus(ctx, host.Account, host.SystemProfile.RHCID)
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
