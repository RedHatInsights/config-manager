package internal

import (
	"config-manager/infrastructure/persistence/cloudconnector"
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

// SetupHost sends a message to the host through cloud-connector to install the
// rhc-worker-playbook RPM, enabling the host to receive and execute playbooks
// sent through playbook-dispatcher.
func SetupHost(ctx context.Context, host Host) (string, error) {
	logger := log.With().Str("org_id", host.OrgID).Str("subscription_manager_id", host.SubscriptionManagerID).Str("rhc_client_id", host.SystemProfile.RHCID).Logger()

	if host.OrgID == "" {
		return "", fmt.Errorf("cannot setup host: missing value for 'orgID' field")
	}

	if host.SystemProfile.RHCID == "" {
		return "", fmt.Errorf("cannot setup host: missing value for 'rhc_client_id' field")
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

		status, dispatchers, err = client.GetConnectionStatus(ctx, host.OrgID, host.SystemProfile.RHCID)
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

	messageID, err := client.SendMessage(ctx, host.OrgID, "package-manager", data, nil, host.SystemProfile.RHCID)
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
		status, dispatchers, err := client.GetConnectionStatus(ctx, host.OrgID, host.SystemProfile.RHCID)
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
