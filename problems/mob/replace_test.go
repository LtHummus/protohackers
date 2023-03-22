package mob

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwapAddress(t *testing.T) {
	assert.Equal(t, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI\n", swapAddress("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX\n"))
	assert.Equal(t, "The product is 7YWHMfk9JZe0LM0g1ZauHuiSxhI-HPVbees8gOSRTzOeroVi1op4tJNoiHr-1234 ok\n", swapAddress("The product is 7YWHMfk9JZe0LM0g1ZauHuiSxhI-HPVbees8gOSRTzOeroVi1op4tJNoiHr-1234 ok\n"))
	assert.Equal(t, "Please pay the ticket price of 15 Boguscoins to one of these addresses: 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI\n", swapAddress("Please pay the ticket price of 15 Boguscoins to one of these addresses: 7WIHZFKXiYyxU7hJmc7MVPSGCPBcL4qFY 7ZksfwZKxJ5DLxaqJMVcDNwqdAjq3Ekm 7uiZHQVtUrkCEoz5BGBWMnUwlbasaM87RDW\n"))
}
