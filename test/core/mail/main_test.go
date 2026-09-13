package mail_test

import (
	"os"
	"testing"

	_ "sumeru/addons/base/models"
	_ "sumeru/addons/mail/models"
	"sumeru/core/modelreg"
	_ "sumeru/core/ormmodels"
)

func TestMain(m *testing.M) {
	if err := modelreg.ActivateAll([]string{"base", "mail"}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
