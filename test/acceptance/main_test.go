package acceptance

import (
	"os"
	"testing"

	_ "sumeru/addons/base/models"
	_ "sumeru/addons/contacts/models"
	_ "sumeru/core/ormmodels"
	"sumeru/core/modelreg"
)

func TestMain(m *testing.M) {
	if err := modelreg.ActivateAll([]string{"base", "contacts"}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
