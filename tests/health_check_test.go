package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckWorks(t *testing.T) {
	testApp := SpawnApp()

	client := &http.Client{}
	response, err := client.Get(testApp.address + "/health")

	assert.NoError(t, err, "Failed to execute request")
	defer response.Body.Close()

	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	assert.Equal(t, int64(0), response.ContentLength)
}
