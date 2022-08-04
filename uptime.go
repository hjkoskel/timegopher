/*
This module provides uptime.

"Algorithm" is based on fact that golang time.Time includes "hidden" monotonic clock that is used when doing time operations like diff.
https://pkg.go.dev/time#hdr-Monotonic_Clocks
So it is possible to resolve what time.Time uptime. Assuming that it happens after latest kernel boot (uptime 0).
For resolving older uptimes from timestamps, other methods and rtc sync history is needed

/proc/uptime have 0.01s granularity meaning that anything happening more frequently than 100Hz will have repeating timestamps

*/
package timegopher

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type UptimeChecker struct {
	//Matching dates. Does not use inaccurate /proc/uptime in every call
	createdUptime NsUptime
	createdTime   time.Time //Get duration, it uses monotonic clock
}

//UptimeNano resolves what is uptime on specific timestamp
func (p *UptimeChecker) UptimeNano(tNow time.Time) (NsUptime, error) {
	if p.createdUptime == 0 { //It can not be 0
		return 0, fmt.Errorf("uptime checker not initialized propely") //Prevents some bugs
	}

	result := p.createdUptime + NsUptime(tNow.Sub(p.createdTime).Nanoseconds())
	if result < 0 {
		return result, fmt.Errorf("time %v is before boot", tNow)
	}
	return result, nil
}

func parseUptimeFile(content []byte) (NsUptime, error) {
	a := strings.Fields(string(content))
	if len(a) != 2 {
		return 0, fmt.Errorf("invalid uptime format %s", content)
	}
	f, errParse := strconv.ParseFloat(a[0], 64)
	if errParse != nil {
		return 0, fmt.Errorf("invalid uptime format %s  (err %v)", content, errParse.Error())
	}
	return NsUptime(f * float64(1000.0*1000.0*1000.0)), nil
}

const (
	PROCUPTIME = "/proc/uptime"
)

//For reference 10ms resolution only
func GetDirectUptime() (NsUptime, error) {
	rawUptime, errRawUptime := ioutil.ReadFile(PROCUPTIME)
	if errRawUptime != nil {
		return 0, errRawUptime
	}
	return parseUptimeFile(rawUptime)
}

//Replace this global variable at tests
var procFS = os.DirFS("/proc")

//Creates uptime checker, reads uptime and sets creation time
func CreateUptimeChecker() (UptimeChecker, error) {
	result := UptimeChecker{}
	//Average uptime.  Trying to be perfectionist :D
	rawUptime0, errRawUptime0 := fs.ReadFile(procFS, "uptime")
	result.createdTime = time.Now()
	rawUptime1, errRawUptime1 := fs.ReadFile(procFS, "uptime")

	if errRawUptime0 != nil {
		return result, errRawUptime0
	}
	if errRawUptime1 != nil {
		return result, errRawUptime0
	}

	ut0, utErr0 := parseUptimeFile(rawUptime0)
	ut1, utErr1 := parseUptimeFile(rawUptime1)
	if utErr0 != nil {
		return result, utErr0
	}
	if utErr1 != nil {
		return result, utErr1
	}

	result.createdUptime = (ut0 + ut1) / 2
	return result, nil
}
