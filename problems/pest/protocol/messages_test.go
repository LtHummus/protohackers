package protocol

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func testSerializeDeserialize(t *testing.T, x Message) {
	b, err := x.Serialize()
	assert.NoError(t, err)

	o, err := Deserialize(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, x, o)
}

func TestSerializeDeserialize(t *testing.T) {
	testObjects := []Message{
		&Hello{
			Protocol: "pestcontrol",
			Version:  1,
		},
		&Error{Message: "this is an error message indicating that something bad happened"},
		&Ok{},
		&DialAuthority{Site: 1985},
		&TargetPopulations{
			Site: 9285,
			Targets: []PopulationTarget{
				{
					Species: "dog",
					Min:     4,
					Max:     5,
				},
				{
					Species: "rat",
					Min:     0,
					Max:     9284,
				},
				{
					Species: "potato",
					Min:     2,
					Max:     9999,
				},
			},
		},
		&CreatePolicy{
			Species: "a really cute cat",
			Action:  Conserve,
		},
		&DeletePolicy{Policy: 12111},
		&PolicyResult{Policy: 9284},
		&SiteVisit{
			Site: 1,
			Observations: []Observation{
				{
					Species: "a calico cat",
					Count:   1,
				},
				{
					Species: "golden retriever",
					Count:   9,
				},
			},
		},
	}

	for _, curr := range testObjects {
		testSerializeDeserialize(t, curr)
	}
}

func TestDeserializeCreatePolicy(t *testing.T) {
	t.Run("fail on invalid policy", func(t *testing.T) {
		b := &CreatePolicy{
			Species: "animal",
			Action:  0x11, // invalid policy
		}

		p, err := b.Serialize()
		require.NoError(t, err)

		_, err = Deserialize(bytes.NewReader(p))
		assert.Error(t, err)
	})
}
