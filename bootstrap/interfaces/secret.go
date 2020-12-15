package interfaces

import "time"

// SecretProvider defines the contract for secret provider implementations that
// allow secrets to be retrieved/stored from/to a services Secret Store.
type SecretProvider interface {
	// StoreSecrets stores new secrets into the service's SecretStore at the specified path.
	StoreSecrets(path string, secrets map[string]string) error

	// GetSecrets retrieves secrets from the service's SecretStore at the specified path.
	GetSecrets(path string, keys ...string) (map[string]string, error)

	// SecretsUpdated sets the secrets last updated time to current time.
	SecretsUpdated()

	// SecretsLastUpdated returns the last time secrets were updated
	SecretsLastUpdated() time.Time
}
