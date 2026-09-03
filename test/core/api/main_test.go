package api_test

import (
	"os"
	"testing"

	_ "sumeru/core/ormmodels"
	"sumeru/core/modelreg"
)

func TestMain(m *testing.M) {
	if err := modelreg.ActivateAll([]string{"base"}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
