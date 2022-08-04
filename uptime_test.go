/*
Sanity test.

Test what locally is available.
*/

package timegopher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicUptime(t *testing.T) {
	tNow := time.Now()

	uptimeNow, errUptime := GetDirectUptime()

	if errUptime != nil {
		t.Errorf("ERROR UPTIME %v", errUptime)
	}

	utChecker, errUtChecker := CreateUptimeChecker()
	if errUtChecker != nil {
		t.Errorf("ERROR creating uptime checker %v", errUtChecker)
	}

	utNano, _ := utChecker.UptimeNano(tNow)
	diffNano := uptimeNow.Diff(utNano)

	if 1000*1000 < diffNano {
		t.Errorf("Something went wrong, direct uptime vs UptimeNano is %vns ", diffNano)
	}

	tNext := tNow.Add(time.Second * 42)
	nextNano, _ := utChecker.UptimeNano(tNext)
	assert.Equal(t, NsUptime(42000000000), nextNano-utNano)
}
