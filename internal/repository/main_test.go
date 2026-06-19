package repository_test

import (
	"fmt"
	"os"
	"testing"

	"Sixth_world_Suday/internal/repository/repotest"
)

func TestMain(m *testing.M) {
	fmt.Println("setup")
	code := m.Run()
	repotest.CleanupTemplate()
	os.Exit(code)
}
