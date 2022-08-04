/*
timegopher_test.go
*/
package timegopher

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColdBoot(t *testing.T) {
	fname := "/tmp/kylma"
	os.Remove(fname)
	cold, coldErr := FirstCallAfterBoot(fname)
	assert.Equal(t, nil, coldErr)
	assert.Equal(t, true, cold)
	cold, coldErr = FirstCallAfterBoot(fname)
	assert.Equal(t, nil, coldErr)
	assert.Equal(t, false, cold)
	os.Remove(fname)

}

//TODO TEST time organizer function per test.. initialize what needed

func TestCreateOrganizerMEM(t *testing.T) {

}
