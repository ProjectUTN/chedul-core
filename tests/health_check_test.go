package tests

import (
	api_test "chedul-core/tests/api"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckWorks(t *testing.T) {
	testApp, err := api_test.SpawnApp(api_test.WithDB())
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
