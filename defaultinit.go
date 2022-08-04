/*
Default time organizer

Helper function that does initialization that is good enough for many use cases
*/
package timegopher

import (
	"fmt"
	"time"

	"github.com/hjkoskel/fixregsto"
)

const (
	DEFAULTDBFILE_UNCERTAINRTC = "uncertain.rtc"
	DEFAULTDBFILE_RTC          = "rtcsync.rtc"
	DEFAULTDBFILE_STARTLOG     = "start.time"
	DEFAULTDBFILE_STOPLOG      = "stop.time"
	DEFAULTDBFILE_ALIVELOG     = "alive.time"
)

const (
	WARMSTARTFILE    = "/tmp/warmstart"
	WARMSTARTCONTENT = "warm start"
)

const (
	RECORDSIZE_TIMEVARIABLE_NORTC = 12
	RECORDSIZE_TIMEVARIABLE_RTC   = 20
)

/*
Create default that is good for embedded linux use
This function acts also as example use
Not possible to unit test well.
*/
func CreateDefaultTimeGopher(rtcLogDir string, latestKnowTimeElsewhere TimeVariable) (TimeGopher, error) {
	var errDisk error

	inSync, errInSync := RtcIsSynced_adjtimex()
	if errInSync != nil {
		return TimeGopher{}, fmt.Errorf("checking rtc sync error= %v", errInSync)
	}

	//var uncertainRtcSyncLog, rtcSyncLog, startupLog, lastLog, volatileAlive, nonVoltatileAlive InDiskDb

	var uncertainRtcLog, rtcLog, startLog, stopLog, lastLog TimeFileDb

	firstRunAfterBoot, errFirstRunAfterBoot := FirstCallAfterBoot(WARMSTARTFILE)
	if errFirstRunAfterBoot != nil {
		return TimeGopher{}, fmt.Errorf("FirstCallAfterBoot:%v", errFirstRunAfterBoot)
	}

	/*DEFAULTDBFILE_UNCERTAINRTC = "uncertain.rtc"
	DEFAULTDBFILE_RTC = "rtcsync.rtc"

	DEFAULTDBFILE_STARTUPLOG = "startup.time"
	DEFAULTDBFILE_ALIVELOG = "alive.time"
	*/

	/*
		Uncertain RTC.
		When user syncs or some unreliable source  "better than nothing"
	*/
	confUncertainRtc := fixregsto.FileStorageConf{
		Name:         DEFAULTDBFILE_UNCERTAINRTC,
		RecordSize:   RECORDSIZE_TIMEVARIABLE_RTC,
		MaxFileCount: 256,
		FileMaxSize:  512 * 4,
		Path:         rtcLogDir,
	}

	stoUncertainRtc, errUncertainRtc := confUncertainRtc.InitFileStorage()
	if errUncertainRtc != nil {
		return TimeGopher{}, fmt.Errorf("UncertainRtc init error %v", errUncertainRtc)
	}
	uncertainRtcLog, errDisk = CreateTimeFileDb(&stoUncertainRtc, true)
	if errDisk != nil {
		return TimeGopher{}, fmt.Errorf("UncertainRTC create error %v", errDisk)
	}

	/*
		Good sync from good clock source (NTP etc...)
	*/
	confRtc := fixregsto.FileStorageConf{
		Name:         DEFAULTDBFILE_RTC,
		RecordSize:   RECORDSIZE_TIMEVARIABLE_RTC,
		MaxFileCount: 256,
		FileMaxSize:  512 * 4,
		Path:         rtcLogDir,
	}

	stoRtc, errRtc := confRtc.InitFileStorage()
	if errRtc != nil {
		return TimeGopher{}, fmt.Errorf("rtc sync init err %v", errRtc)
	}
	rtcLog, errDisk = CreateTimeFileDb(&stoRtc, true)
	if errDisk != nil {
		return TimeGopher{}, fmt.Errorf("rtc create error %v", errDisk)
	}

	/*
		Start log,

		Updated when program starts (copies previous alive)
	*/
	confStart := fixregsto.FileStorageConf{
		Name:         DEFAULTDBFILE_STARTLOG,
		RecordSize:   RECORDSIZE_TIMEVARIABLE_NORTC,
		MaxFileCount: 256,
		FileMaxSize:  512 * 4,
		Path:         rtcLogDir,
	}

	stoStart, errStartLast := confStart.InitFileStorage()
	if errStartLast != nil {
		return TimeGopher{}, fmt.Errorf("startlog init err %v", errStartLast)
	}
	startLog, errDisk = CreateTimeFileDb(&stoStart, false)
	if errDisk != nil {
		return TimeGopher{}, fmt.Errorf("startlog create err %v", errDisk)
	}

	/*
		STOP
	*/
	confStop := fixregsto.FileStorageConf{
		Name:         DEFAULTDBFILE_STOPLOG,
		RecordSize:   RECORDSIZE_TIMEVARIABLE_NORTC,
		MaxFileCount: 256,
		FileMaxSize:  512 * 4,
		Path:         rtcLogDir,
	}

	stoStop, errStopLast := confStop.InitFileStorage()
	if errStopLast != nil {
		return TimeGopher{}, errStartLast
	}
	stopLog, errDisk = CreateTimeFileDb(&stoStop, false)
	if errDisk != nil {
		return TimeGopher{}, errDisk
	}

	//****** Alive log. Only few entries needed
	confAlive := fixregsto.FileStorageConf{
		Name:         DEFAULTDBFILE_ALIVELOG,
		RecordSize:   RECORDSIZE_TIMEVARIABLE_NORTC,
		MaxFileCount: 1,   //At least one, so no "no points" situation can happen when work flushes
		FileMaxSize:  512, //Todo prefered min size somewhere?
		Path:         rtcLogDir,
	}

	stoLast, errLast := confAlive.InitFileStorage()
	if errLast != nil {
		return TimeGopher{}, errLast
	}
	lastLog, errDisk = CreateTimeFileDb(&stoLast, false)
	if errDisk != nil {
		return TimeGopher{}, errDisk
	}

	uptimeCheck, errCreateUptimeChecker := CreateUptimeChecker()
	if errCreateUptimeChecker != nil {
		return TimeGopher{}, errCreateUptimeChecker
	}

	result, newErr := NewTimeGopher(
		time.Now(), //timeNow time.Time,
		inSync,
		firstRunAfterBoot,
		//These have RTC time
		&rtcLog,
		&uncertainRtcLog,
		&startLog,
		&stopLog,
		&lastLog,
		latestKnowTimeElsewhere, // TimeVariable, //If knows from latest stored timestamp on timeseries database
		&uptimeCheck,
	)
	if newErr != nil {
		return result, fmt.Errorf("NewTimeGopher error %v", newErr)
	}
	return result, nil
}
