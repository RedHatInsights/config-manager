package authorization

import "context"

func getDefaultWorkspaceID(context context.Context, orgID string) (string, error) {

	//
	// TODO
	// - add configuration options for rbac (host, psk)
	// - populate rbac host from clowder configuration
	// - make call to http://<rbac>/api/rbac/v2/workspaces/?type=default&limit=1
	// - pass in the rbac PSK as a request header
	// - parse the workspace id from the response and return
	// - add metrics and logs for the rbac invocation

	return "", nil
}
