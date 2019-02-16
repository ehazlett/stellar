package stellar

const (
	// APIVersion is the version of the API
	APIVersion = "v1"
	// StellarNetworkLabel is the label to identify that a container should use stellar networking
	StellarNetworkLabel = "stellar.io/network"
	// StellarApplicationLabel is the label to identify that a container belongs to the stellar app
	StellarApplicationLabel = "stellar.io/application"
	// StellarRestartLabel specifies that the service should be restarted on stop or failure
	StellarRestartLabel     = "stellar.io/restart"
	StellarExtensionID      = "stellar.io/extensions"
	StellarServiceExtension = StellarExtensionID + "/Service"
)
