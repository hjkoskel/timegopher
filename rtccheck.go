/*
Helper functions for checking RTC status.

Checking is linux wall clock in sync depends on installation and hardware configuration.
Initial guess is that Adjtimex should work.

Please add functions for checking synchronization status in some other methods

*/
package timegopher

import (
	"syscall"
)

// https://man7.org/linux/man-pages/man2/adjtimex.2.html
const (
	TIME_OK   = iota //Clock synchronized, no leap second adjustment pending.
	TIME_INS         //Indicates that a leap second will be added at the end of the UTC day
	TIME_DEL         //Indicates that a leap second will be deleted at the end of the UTC day.
	TIME_OOP         //Insertion of a leap second is in progress.
	TIME_WAIT        // A leap-second insertion or deletion has been completed. This value will be returned until the next ADJ_STATUS operation clears the STA_INS and STA_DEL flags.
	TIME_ERROR
	/*
		TIME_ERROR
			The system clock is not synchronized to a reliable server.
			              This value is returned when any of the following holds
			              true:

			              *  Either STA_UNSYNC or STA_CLOCKERR is set.

			              *  STA_PPSSIGNAL is clear and either STA_PPSFREQ or
			                 STA_PPSTIME is set.

			              *  STA_PPSTIME and STA_PPSJITTER are both set.

			              *  STA_PPSFREQ is set and either STA_PPSWANDER or
			                 STA_PPSJITTER is set.

			              The symbolic name TIME_BAD is a synonym for TIME_ERROR,
			              provided for backward compatibility.
	*/
)

const (
	STA_PLL       = 0x0001 /* enable PLL updates (rw) */
	STA_PPSFREQ   = 0x0002 /* enable PPS freq discipline (rw) */
	STA_PPSTIME   = 0x0004 /* enable PPS time discipline (rw) */
	STA_FLL       = 0x0008 /* select frequency-lock mode (rw) */
	STA_INS       = 0x0010 /* insert leap (rw) */
	STA_DEL       = 0x0020 /* delete leap (rw) */
	STA_UNSYNC    = 0x0040 /* clock unsynchronized (rw) */
	STA_FREQHOLD  = 0x0080 /* hold frequency (rw) */
	STA_PPSSIGNAL = 0x0100 /* PPS signal present (ro) */
	STA_PPSJITTER = 0x0200 /* PPS signal jitter exceeded (ro) */
	STA_PPSWANDER = 0x0400 /* PPS signal wander exceeded (ro) */
	STA_PPSERROR  = 0x0800 /* PPS signal calibration error (ro) */
	STA_CLOCKERR  = 0x1000 /* clock hardware fault (ro) */
	STA_NANO      = 0x2000 /* resolution (0 = us, 1 = ns) (ro) */
	STA_MODE      = 0x4000 /* mode (0 = PLL, 1 = FLL) (ro) */
	STA_CLK       = 0x8000 /* clock source (0 = A, 1 = B) (ro) */
)

//RtcIsSynced_adjtimex uses syscall.Adjtimex for checking is wall clock synchronized
func RtcIsSynced_adjtimex() (bool, error) {
	tx := syscall.Timex{}
	rtcState, err := syscall.Adjtimex(&tx)
	if err != nil {
		return false, err
	}

	return (rtcState <= 0) && (rtcState != TIME_ERROR), nil
}
