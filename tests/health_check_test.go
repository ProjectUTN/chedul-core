package tests

import (
	"chedul-core/tests/api"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckWorks(t *testing.T) {
	configPath := os.Getenv("CONFIG_DIR")
	if configPath == "" {
		t.Skip("env var 'CONFIG_DIR' no inicializada")
	}

	testApp, err := api.SpawnApp()
	if err != nil {
		log.Fatal(err)
	}

	defer testApp.Cleanup()

	client := &http.Client{}
	response, err := client.Get(testApp.Address + "/health")

	assert.NoError(t, err, "Failed to execute request")
	defer response.Body.Close()

	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	assert.Equal(t, int64(0), response.ContentLength)
}
