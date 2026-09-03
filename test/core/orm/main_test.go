package orm_test

import (
	"os"
	"testing"

	_ "sumeru/addons/base/models"
	_ "sumeru/addons/contacts/models"
	"sumeru/core/modelreg"
	_ "sumeru/core/ormmodels"
)

func TestMain(m *testing.M) {
	if err := modelreg.ActivateAll([]string{"base", "contacts"}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
