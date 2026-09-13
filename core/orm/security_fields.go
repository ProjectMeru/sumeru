package orm

import "sumeru/core/security"

func readRedactFields(model string) map[string]bool {
	return security.ReadRedactFields(model)
}

func writeDenyUnlessSysFields(model string) map[string]bool {
	return security.WriteDenyUnlessSysFields(model)
}

func writeDenyDirectFields(model string) map[string]bool {
	return security.WriteDenyDirectFields(model)
}
