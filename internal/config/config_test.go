package config

import (
	"testing"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

func TestResolveRbacURL_V2WithCA(t *testing.T) {
	ca := "/tmp/ca-bundle.crt"
	clowder.DependencyEndpointsV2 = map[string]map[string]clowder.DependencyEndpointV2{
		"rbac": {
			"service": {Uri: "https://rbac.svc:8443", Authenticated: true, CaCertificate: &ca},
		},
	}
	defer func() { clowder.DependencyEndpointsV2 = nil }()

	url, caCert := resolveRbacURL()
	if url != "https://rbac.svc:8443" {
		t.Errorf("expected https://rbac.svc:8443, got %s", url)
	}
	if caCert != "/tmp/ca-bundle.crt" {
		t.Errorf("expected /tmp/ca-bundle.crt, got %s", caCert)
	}
}

func TestResolveRbacURL_V2WithoutCA(t *testing.T) {
	clowder.DependencyEndpointsV2 = map[string]map[string]clowder.DependencyEndpointV2{
		"rbac": {
			"service": {Uri: "http://rbac-service.rbac.svc:8000", Authenticated: false},
		},
	}
	defer func() { clowder.DependencyEndpointsV2 = nil }()

	url, caCert := resolveRbacURL()
	if url != "http://rbac-service.rbac.svc:8000" {
		t.Errorf("expected http://rbac-service.rbac.svc:8000, got %s", url)
	}
	if caCert != "" {
		t.Errorf("expected empty caCert, got %s", caCert)
	}
}

func TestResolveRbacURL_V2EmptyURI(t *testing.T) {
	// V2 present but URI empty — should fall through to V1
	clowder.DependencyEndpointsV2 = map[string]map[string]clowder.DependencyEndpointV2{
		"rbac": {
			"service": {Uri: "", Authenticated: false},
		},
	}
	clowder.LoadedConfig = &clowder.AppConfig{
		Endpoints: []clowder.DependencyEndpoint{
			{App: "rbac", Hostname: "rbac-v1.svc", Port: 8080},
		},
	}
	defer func() {
		clowder.DependencyEndpointsV2 = nil
		clowder.LoadedConfig = nil
	}()

	url, caCert := resolveRbacURL()
	if url != "http://rbac-v1.svc:8080" {
		t.Errorf("expected http://rbac-v1.svc:8080, got %s", url)
	}
	if caCert != "" {
		t.Errorf("expected empty caCert, got %s", caCert)
	}
}

func TestResolveRbacURL_V1Fallback(t *testing.T) {
	clowder.DependencyEndpointsV2 = nil
	clowder.LoadedConfig = &clowder.AppConfig{
		Endpoints: []clowder.DependencyEndpoint{
			{App: "host-inventory", Hostname: "inventory.svc", Port: 8000},
			{App: "rbac", Hostname: "rbac-host.svc", Port: 8000},
		},
	}
	defer func() { clowder.LoadedConfig = nil }()

	url, caCert := resolveRbacURL()
	if url != "http://rbac-host.svc:8000" {
		t.Errorf("expected http://rbac-host.svc:8000, got %s", url)
	}
	if caCert != "" {
		t.Errorf("expected empty caCert, got %s", caCert)
	}
}

func TestResolveRbacURL_NoEndpoints(t *testing.T) {
	clowder.DependencyEndpointsV2 = nil
	clowder.LoadedConfig = &clowder.AppConfig{
		Endpoints: []clowder.DependencyEndpoint{},
	}
	defer func() { clowder.LoadedConfig = nil }()

	url, caCert := resolveRbacURL()
	if url != "" {
		t.Errorf("expected empty url, got %s", url)
	}
	if caCert != "" {
		t.Errorf("expected empty caCert, got %s", caCert)
	}
}

func TestResolveRbacURL_V2NilCA(t *testing.T) {
	// V2 endpoint with nil CaCertificate — should work, no CA cert
	clowder.DependencyEndpointsV2 = map[string]map[string]clowder.DependencyEndpointV2{
		"rbac": {
			"service": {Uri: "http://rbac.svc:8000", Authenticated: false, CaCertificate: nil},
		},
	}
	defer func() { clowder.DependencyEndpointsV2 = nil }()

	url, caCert := resolveRbacURL()
	if url != "http://rbac.svc:8000" {
		t.Errorf("expected http://rbac.svc:8000, got %s", url)
	}
	if caCert != "" {
		t.Errorf("expected empty caCert, got %s", caCert)
	}
}

func TestResolveRbacURL_V2WrongApp(t *testing.T) {
	// V2 present but for different app — fall through to V1
	clowder.DependencyEndpointsV2 = map[string]map[string]clowder.DependencyEndpointV2{
		"sources-api": {
			"svc": {Uri: "http://sources.svc:8000", Authenticated: false},
		},
	}
	clowder.LoadedConfig = &clowder.AppConfig{
		Endpoints: []clowder.DependencyEndpoint{
			{App: "rbac", Hostname: "rbac-v1.svc", Port: 9000},
		},
	}
	defer func() {
		clowder.DependencyEndpointsV2 = nil
		clowder.LoadedConfig = nil
	}()

	url, caCert := resolveRbacURL()
	if url != "http://rbac-v1.svc:9000" {
		t.Errorf("expected http://rbac-v1.svc:9000, got %s", url)
	}
	if caCert != "" {
		t.Errorf("expected empty caCert, got %s", caCert)
	}
}

func TestResolveRbacURL_V2EmptyCAString(t *testing.T) {
	// V2 endpoint with empty string CaCertificate — should not set CA
	empty := ""
	clowder.DependencyEndpointsV2 = map[string]map[string]clowder.DependencyEndpointV2{
		"rbac": {
			"service": {Uri: "http://rbac.svc:8000", Authenticated: true, CaCertificate: &empty},
		},
	}
	defer func() { clowder.DependencyEndpointsV2 = nil }()

	url, caCert := resolveRbacURL()
	if url != "http://rbac.svc:8000" {
		t.Errorf("expected http://rbac.svc:8000, got %s", url)
	}
	if caCert != "" {
		t.Errorf("expected empty caCert, got %s", caCert)
	}
}
