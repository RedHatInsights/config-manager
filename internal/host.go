package internal

import (
	"config-manager/infrastructure/persistence/cloudconnector"
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// Host represents a system record from the Inventory application.
type Host struct {
	ID            string `json:"id"`
	Account       string `json:"account"`
	OrgID         string `json:"org_id"`
	DisplayName   string `json:"display_name"`
	Reporter      string `json:"reporter"`
	SystemProfile struct {
		RHCID    string `json:"rhc_client_id"`
		RHCState string `json:"rhc_config_state"`
	} `json:"system_profile"`
}

// ApplyProfile applies the current profile to the specified hosts.
func ApplyProfile(ctx context.Context, profile *db.Profile, hosts []Host) ([]dispatcher.RunCreated, error) {
	var err error
	var results []dispatcher.RunCreated
	var inputs []dispatcher.RunInput

	if !profile.Active {
		log.Info().Msgf("account_state.apply_state is false; skipping configuration")
		return []dispatcher.RunCreated{}, nil
	}

	log.Info().Interface("state", profile.StateConfig()).Interface("clients", hosts).Msgf("start applying state")
	for i, client := range hosts {
		logger := log.With().Str("client_id", client.SystemProfile.RHCID).Interface("client", client).Logger()

		if client.Reporter == "cloud-connector" {
			logger.Debug().Msg("setting up host for playbook execution")
			if _, err := SetupHost(context.Background(), client); err != nil {
				logger.Error().Err(err).Msg("cannot set up host for playbook execution")
				continue
			}
		}

		logger.Info().Msg("dispatching work for client")
		input := dispatcher.RunInput{
			Recipient: client.SystemProfile.RHCID,
			Account:   profile.AccountID.String,
			Url:       config.DefaultConfig.PlaybookHost.String() + fmt.Sprintf(config.DefaultConfig.PlaybookPath, profile.ID),
			Labels: &dispatcher.RunInput_Labels{
				AdditionalProperties: map[string]string{
					"state_id": profile.ID.String(),
					"id":       client.ID,
				},
			},
		}
		logger.Debug().Interface("run_input", input).Msg("created run input")

		inputs = append(inputs, input)

		if len(inputs) == config.DefaultConfig.DispatcherBatchSize || i == len(hosts)-1 {
			if inputs != nil {
				logger.Debug().Interface("inputs", inputs).Msg("dispatching runs to playbook-dispatcher")
				client := dispatcher.NewDispatcherClient()
				res, err := client.Dispatch(ctx, inputs)
				if err != nil {
					logger.Error().Err(err).Msg("cannot dispatch work to playbook dispatcher - giving up")
					continue
				}

				results = append(results, res...)
				inputs = nil
				logger.Debug().Interface("results", results).Msg("results from dispatch")
			}
		}
	}
	log.Info().Msg("finish applying state")

	return results, err
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
