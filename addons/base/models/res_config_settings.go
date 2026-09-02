package models

import (
	"sumeru/core/sdk"
)

// ResConfigSettings is the platform settings shell extended by application modules.
type ResConfigSettings struct {
	sdk.Model `sumeru:"model=res.config.settings"`
}
