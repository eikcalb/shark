package constants

type ServiceContextKey string

const (
	CONTEXT_APPLICATION_VERSION_KEY ServiceContextKey = "CONTEXT_APPLICATION_VERSION_KEY"
	CONTEXT_SERVICE_VERSION_KEY     ServiceContextKey = "CONTEXT_SERVICE_VERSION_KEY"
	CONTEXT_SERVICE_PORT_KEY        ServiceContextKey = "CONTEXT_SERVICE_PORT_KEY"
)
