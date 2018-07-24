package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var fakeTargetAddress = "0x0A6d6733Cf17311499184aB509E4590A09952ba4"

func TestGetEtherAccountIsInvalid(t *testing.T) {
	account, err := getEtherAddress("/send some_invalid_address")
	assert.Nil(t, account)
	assert.Error(t, ErrEtherAddressInvalid, err)

	account, err = getEtherAddress("/send 0x123412342")
	assert.Nil(t, account)
	assert.Error(t, ErrEtherAddressInvalid, err)
}

func TestGetEtherAccountCommandInvalid(t *testing.T) {
	account, err := getEtherAddress("some junk text")
	assert.Nil(t, account)
	assert.Error(t, ErrCommandInvalid, err)

	account, err = getEtherAddress("")
	assert.Nil(t, account)
	assert.Error(t, ErrCommandInvalid, err)
}

func TestGetEtherAccountCommandIncomplete(t *testing.T) {
	account, err := getEtherAddress("/send ")
	assert.Nil(t, account)
	assert.Error(t, ErrCommandIncomplete, err)
}

func TestGetEtherAccountAddressValid(t *testing.T) {
	address, err := getEtherAddress("/send 0x0A6d6733Cf17311499184aB509E4590A09952ba4")
	assert.NoError(t, err)
	assert.Equal(t, address.String(), fakeTargetAddress)

	address, err = getEtherAddress("/send   0x0A6d6733Cf17311499184aB509E4590A09952ba4")
	assert.NoError(t, err)
	assert.Equal(t, address.String(), fakeTargetAddress)

	address, err = getEtherAddress("/send 0x0A6d6733Cf17311499184aB509E4590A09952ba4   ")
	assert.NoError(t, err)
	assert.Equal(t, address.String(), fakeTargetAddress)
}
