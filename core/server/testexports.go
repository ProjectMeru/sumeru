package server

import "sumeru/core/server/config"

// ListenAddrForTest exposes listenAddr for tests.
func ListenAddrForTest(host, port string) string {
	return listenAddr(host, port)
}

// SetupListenAddrForTest exposes setupListenAddr for tests.
func SetupListenAddrForTest(cfg config.Config) string {
	return setupListenAddr(cfg)
}

// ValidateSetupModeConfigForTest exposes validateSetupModeConfig for tests.
func ValidateSetupModeConfigForTest(cfg config.Config) error {
	return validateSetupModeConfig(cfg)
}

// ConfigForTest builds a minimal config for listen tests.
func ConfigForTest(httpInterface, httpPort string, setupLocalhostOnly bool) config.Config {
	return config.Config{
		HttpInterface:      httpInterface,
		HttpPort:           httpPort,
		SetupLocalhostOnly: setupLocalhostOnly,
	}
}

// ConfigForSetupTest builds config with setup token for listen/validation tests.
func ConfigForSetupTest(httpInterface, httpPort, setupToken string, setupLocalhostOnly bool) config.Config {
	return config.Config{
		HttpInterface:      httpInterface,
		HttpPort:           httpPort,
		SetupToken:         setupToken,
		SetupLocalhostOnly: setupLocalhostOnly,
	}
}

// NormalizeManifestAssetRelForTest exposes manifest asset path normalization for tests.
func NormalizeManifestAssetRelForTest(rel string) (string, bool) {
	return normalizeManifestAssetRel(rel)
}

// ManifestAssetPublicURLForTest builds the public URL for a manifest asset entry.
func ManifestAssetPublicURLForTest(moduleName, cleanRel string) string {
	return manifestAssetPublicURL(moduleName, cleanRel)
}
