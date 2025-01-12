/*
Timesync library for syncing from external sources like NTP
*/
package timesync

import "time"

type TimeSync interface {
	//Get difference to time now
	GetOffset() (time.Duration, error)
}
