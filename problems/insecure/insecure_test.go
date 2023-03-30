package insecure

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPopularToy(t *testing.T) {
	assert.Equal(t, "15x dog on a string", getPopularToy("10x toy car,15x dog on a string,4x inflatable motorcycle"))
	assert.Equal(t, "5x string", getPopularToy("5x string"))
	assert.Equal(t, "10x pony", getPopularToy("10x pony,10x horse"))
}
