/*
TimeGopher

AJATUS 18.7.2022
Talletetaan RTC kellosync mukaan jos sync


AJATUS 19.7.2022
Voisiko joitain funktioita laittaa erikseen? Jotta helpompi testattavuus
*/
package timegopher

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type TimeGopher struct {
	synced bool //Is good RTC known.

	RtcMaxDeviation NsEpoch //deviation in nanosecond from RTC when in sync. Set variable default value if need to change settings

	UncertainRtcSyncLog *TimeFileDb //For manual sync
	RtcSyncLog          *TimeFileDb
	StartLog            *TimeFileDb //boot number and uptime
	StopLog             *TimeFileDb //boot number and uptime needed
	LastLog             *TimeFileDb //Last alive situation

	coldStart bool //VolatileAlive     *TimeFileDb //Detects is there resets,

	//Last item on start log BootNumber int32
	bootNumber int32

	UptimeCheck *UptimeChecker //Create externally, better for testing
}

//GetLatestTime picks the last entry of any TimeFileDb entry inside TimeGopher instance. Used internally and for diagnostics
func (p *TimeGopher) GetLatestTime() (TimeVariable, error) {
	result := TimeVariable{BootNumber: 0}

	dbArr := []*TimeFileDb{
		p.RtcSyncLog,
		p.StartLog,
		p.StopLog,
		p.LastLog,
		p.UncertainRtcSyncLog,
	}

	for _, db := range dbArr {
		if db == nil {
			continue
		}
		arr, errArr := (*db).GetLatestN(1)
		if errArr != nil {
			return result, errArr
		}
		if 0 < len(arr) {
			if result.BootNumber < arr[0].BootNumber {
				result = arr[0]
			}
			if result.BootNumber == arr[0].BootNumber {
				if result.Uptime < arr[0].Uptime { //Same boot, compare with uptime
					result = arr[0]
				}
			}
		}
	}
	return result, nil
}

//FirstStartAfterBoot, helper function. Call and tell is this the first time. Creates file.
func FirstCallAfterBoot(flagfilename string) (bool, error) {
	info, err := os.Stat(flagfilename)
	if os.IsNotExist(err) {
		e := ioutil.WriteFile(flagfilename, []byte(WARMSTARTCONTENT), 0666)
		return true, e
	}
	if info.IsDir() {
		return false, fmt.Errorf("file %v is dir", flagfilename)
	}
	return false, nil
}

//NewTimeGopher initializes TimeGopher
//Call only once per software run. If this is too complicated and customization is needed then call CreateDefaultTimeGopher( instead.

//Parameters:
//	timeNow, give time.Now() as parameter
//	inSync, set true if system time is synchronized when TimeGopher is created. Resolve for example with RtcIsSynced_adjtimex()
//	coldStart, set true if first start. Resolve this for example with FirstCallAfterBoot(WARMSTARTFILE)
//	rtcSyncLog, TimeFileDb pointer for storing certain sync events
//	uncertainRtcSyncLog, TimeFileDb pointer for storing uncertain sync events. Nil if not needed
//	startLog TimeFileDb, TimeFileDb pointer for storing entries when software starts (colds and warms). Nil if not needed
//	stopLog TimeFileDb, TimeFileDb pointer for storing entries when software stops (entries added at next TimeGopher init). Nil if not needed
//	lastLog TimeFileDb, TimeFileDb pointer for keeping up situation status when sofware was running
//	latestKnowTimeElsewhere TimeVariable, //If some other time stamp information is kept outside TimeGopher, get latest entry here
//	uptimeCheck, Pointer for uptime checker. There can be many implementations depeding on needs. (or unit test requires dummy version)
func NewTimeGopher(
	timeNow time.Time,

	inSync bool,
	coldStart bool,
	//These have RTC time
	rtcSyncLog *TimeFileDb,
	uncertainRtcSyncLog *TimeFileDb, //Optional

	startLog *TimeFileDb, //Optional
	stopLog *TimeFileDb, //Optional

	lastLog *TimeFileDb, //Last alive situation

	latestKnowTimeElsewhere TimeVariable, //If knows from latest stored timestamp on timeseries database
	uptimeCheck *UptimeChecker,
) (TimeGopher, error) {

	result := TimeGopher{
		synced:              inSync,
		RtcMaxDeviation:     5 * 1000 * 1000 * 1000, //TODO ADD AS PARAMETER. Or change default separately
		UncertainRtcSyncLog: uncertainRtcSyncLog,
		RtcSyncLog:          rtcSyncLog,
		StartLog:            startLog,
		StopLog:             stopLog,
		LastLog:             lastLog,

		coldStart:   coldStart,
		UptimeCheck: uptimeCheck,
	}

	if result.RtcSyncLog == nil {
		return result, fmt.Errorf("RtcSyncLog required")
	}

	latestTime, errBoot := result.GetLatestTime()
	if errBoot != nil {
		return result, fmt.Errorf("NewTimeGopher failed getting latest time err=%v", errBoot.Error())
	}
	if latestKnowTimeElsewhere.After(latestTime) {
		latestTime = latestKnowTimeElsewhere
	}
	//Record latest to stoplog IF needed
	if result.StopLog != nil && 0 < latestTime.Uptime {
		errInsertStop := (*result.StopLog).Insert(latestTime)
		if errInsertStop != nil {
			return result, fmt.Errorf("NewTimeGopher failed inserting %#v", errInsertStop.Error())
		}
	}

	//Determine bootNumber. Cold boot means increase in boot counter
	result.bootNumber = latestTime.BootNumber
	if result.coldStart {
		result.bootNumber++
	}

	//Insert RTC sync if synced but log does not have sync entry

	//Add uncertain. At least one instead on going 1970's   IF it is cold start then entry is required
	if !result.synced {
		tNow, errTNow := result.Convert(timeNow)
		tNow.Epoch = NsEpoch(timeNow.UnixNano()) //Insert bad guess, better than nothing
		if errTNow != nil {
			return result, fmt.Errorf("converting timeNow=%v to TimeVariable fail %v", timeNow, errTNow)
		}

		if result.coldStart {
			insertErr := (*result.UncertainRtcSyncLog).Insert(tNow) //At least one
			if insertErr != nil {
				return result, fmt.Errorf("error inserting uncertainRTCSyncLog at init err=%v", insertErr)
			}
		} else {
			n, errN := (*result.UncertainRtcSyncLog).Len()
			if errN != nil {
				return result, fmt.Errorf("UncertainRtcSyncLog len err %v", errN)
			}
			if n == 0 { //In theory could not happen if warm start. Except if file is lost?
				insertErr := (*result.UncertainRtcSyncLog).Insert(tNow) //At least one
				if insertErr != nil {
					return result, fmt.Errorf("error inserting UncertainRtcSyncLog %v", insertErr.Error())
				}
			}
		}
	}

	//Record latest and refresh
	ut, errUt := result.UptimeCheck.UptimeNano(timeNow)
	if errUt != nil {
		return result, fmt.Errorf("UptimeCheck fail %v", errUt.Error())
	}

	//Recod startLog
	if result.StartLog != nil {
		errStartInsert := (*result.StartLog).Insert(TimeVariable{BootNumber: result.bootNumber, Uptime: NsUptime(ut)})
		if errStartInsert != nil {
			return result, fmt.Errorf("error inserting start %v", errStartInsert.Error())
		}
	}

	errRefresh := result.Refresh(timeNow, result.synced)
	if errRefresh != nil {
		return result, fmt.Errorf("error refreshing NewTimeGopher err=%v", errRefresh.Error())
	}
	return result, nil
}

//IsColdStart() returns true if TimeOrganized have created at first time after boot
//One use for this function is for checking, is there need to do something "after boot" on system
func (p *TimeGopher) IsColdStart() bool {
	return p.coldStart
}

//DoUncertainTimeSyncNow is helper function for calling DoUncertainTimeSync
func (p *TimeGopher) DoUncertainTimeSyncNow() error {
	return p.DoUncertainTimeSync(time.Now())
}

//UncertainTimeSync called by library user, after realtime clock is set from unreliable source like set manually
//This function adds time to uncertain RTC sync log. Uncertain sync is used if certain sync is not available
func (p *TimeGopher) DoUncertainTimeSync(t time.Time) error {
	if p.UncertainRtcSyncLog == nil {
		return fmt.Errorf("uncertain RTC sync log is not set")
	}

	tNow, tNowErr := p.Convert(t)
	if tNowErr != nil {
		return tNowErr
	}
	tNow.Epoch = NsEpoch(t.UnixNano()) //Insert bad guess, better than nothing
	err := (*p.UncertainRtcSyncLog).Insert(tNow)
	if err != nil {
		return err
	}
	return nil
}

//RefreshNow is helper function for Refresh
func (p *TimeGopher) RefreshNow() error {
	inSync, errInSync := RtcIsSynced_adjtimex()
	if errInSync != nil {
		return fmt.Errorf("RefreshNow checking rtc sync error= %v", errInSync)
	}
	return p.Refresh(time.Now(), inSync)
}

//Refresh function is called as often as application requires.
//Calling frequently creates frequent synclog entries so determining when sofware was running
func (p *TimeGopher) Refresh(t time.Time, inSync bool) error {
	tNow, errTNow := p.Convert(t)
	if errTNow != nil {
		return fmt.Errorf("Convert error %v at Refresh", errTNow.Error())
	}

	tNow.Epoch = NsEpoch(t.UnixNano()) //Needed because convert time might set epoch if epoch sync was not found

	if inSync {
		if p.synced { //In sync and still says that it is. Can drift thou
			arrLatest, errArrLatest := (p.RtcSyncLog).GetLatestN(1)
			if errArrLatest != nil {
				return errArrLatest
			}

			needFresh := len(arrLatest) == 0
			if 0 < len(arrLatest) {
				if arrLatest[0].BootNumber < p.bootNumber {
					needFresh = true
				}
				drift, driftErr := p.RtcDeviation(t)
				if driftErr != nil {
					return fmt.Errorf("error getting drift error %v", driftErr.Error())
				}
				if p.RtcMaxDeviation < drift {
					needFresh = true
				}
			}

			if needFresh {
				err := (*p.RtcSyncLog).Insert(tNow)
				if err != nil {
					return err
				}
			}
		} else { //State changed to sync
			err := (*p.RtcSyncLog).Insert(tNow)
			if err != nil {
				return err
			}
		}
	}

	p.synced = inSync

	if p.LastLog != nil {
		err := (*p.LastLog).Insert(tNow)
		if err != nil {
			return fmt.Errorf("inserting %#v failed with err=%#v", tNow, err)
		}
	}
	return nil
}

//Unconvert converts TimeVariable to time.Time, vased on what is synchronization is added. Helper function for SolveTime
func (p *TimeGopher) Unconvert(tv TimeVariable) (time.Time, error) {
	return p.SolveTime(tv.BootNumber, tv.Uptime)
}

//Convert time at current boot to TimeVariable
func (p *TimeGopher) Convert(t time.Time) (TimeVariable, error) {
	ut, utCheckErr := p.UptimeCheck.UptimeNano(t)
	if utCheckErr != nil {
		return TimeVariable{}, utCheckErr
	}
	//Easy case, at this boot. Synced wall clock time is not needed now
	if 0 <= ut {
		result := TimeVariable{
			BootNumber: p.bootNumber, //TODO CHANGE, is this at really at this boot
			Uptime:     NsUptime(ut),
		}
		if p.synced {
			result.Epoch = NsEpoch(t.UnixNano())
		}
		return result, nil
	}
	//Have to search on what boot this might happend. Variable t must have synced wall clock time
	result := TimeVariable{Epoch: NsEpoch(t.UnixNano())}
	if result.Epoch < EPOCH70S {
		return result, fmt.Errorf("Convert:Missing epoch, 1970's not supported")
	}
	//search start boot at first

	rtcBoot, _ := p.RtcSyncLog.SolveBootNumber(result.Epoch)
	//skip error checking. Let fail at SolveUptime
	tvRtc, _ := p.RtcSyncLog.SearchTimeVariable(result.Epoch)

	var utUcResult NsUptime
	var tvUcRtc TimeVariable
	utUcResultErr := fmt.Errorf("not determined")
	if p.UncertainRtcSyncLog != nil {
		rtcUcBoot, _ := p.UncertainRtcSyncLog.SolveBootNumber(result.Epoch)
		//Best uncertain entry is the last entry if time is configured manually (user fix errors)
		ucTimeArr, errUcTimeArr := p.UncertainRtcSyncLog.GetOnBoot(rtcUcBoot)
		tvUcRtc = TimeVariable{BootNumber: -1}
		if errUcTimeArr == nil && 0 < len(ucTimeArr) {
			tvUcRtc = ucTimeArr[len(ucTimeArr)-1]
		}
		utUcResult, utUcResultErr = tvUcRtc.SolveUptime(rtcUcBoot, result.Epoch)

	}
	//Both possible, compare and choose "best". Higher boot number? Just choose synced if both at same boot
	//TODO: add way to choose style how to resolve. Might depend on application

	utResult, utResultErr := tvRtc.SolveUptime(rtcBoot, result.Epoch)

	if utUcResultErr != nil && utResultErr != nil {
		return TimeVariable{}, fmt.Errorf("rtcsynclog solve fail =\"%s\" and uncertain fail=\"%s\"", utResultErr.Error(), utUcResultErr.Error())
	}
	if utUcResultErr == nil && utResultErr != nil {
		result.Uptime = utUcResult
		return result, nil
	}
	if utUcResultErr != nil && utResultErr == nil {
		result.Uptime = utResult
		return result, nil
	}
	//Both
	if tvUcRtc.BootNumber <= tvRtc.BootNumber {
		result.Uptime = utResult
		return result, nil
	}
	result.Uptime = utUcResult
	return result, nil
}

//SolveTime converts boot number and uptime to golang time.Time
func (p *TimeGopher) SolveTime(boot int32, uptime NsUptime) (time.Time, error) {
	epoch, epochErr := p.RtcSyncLog.SolveEpoch(boot, uptime)

	var epochUc NsEpoch
	epochUcErr := fmt.Errorf("not solved")
	if p.UncertainRtcSyncLog != nil {
		epochUc, epochUcErr = p.UncertainRtcSyncLog.SolveEpoch(boot, uptime)
	}
	if epochErr == nil {
		return time.Unix(0, int64(epoch)), nil
	}

	if epochErr != nil && epochUcErr != nil {
		return time.Unix(0, 0), fmt.Errorf("SolveTime err=%v and uncertain synclog err =%v", epochErr.Error(), epochUcErr.Error())
	}
	return time.Unix(0, int64(epochUc)), nil

}

//RtcDeviation gets RTC deviation now. Deviation can happen if timekeeping jumps Based on this, decide is RTC update needed
func (p *TimeGopher) RtcDeviation(t time.Time) (NsEpoch, error) {
	refVar, errRef := p.Convert(t)
	if errRef != nil {
		return NsEpoch(0), errRef
	}

	tv, tvErr := p.RtcSyncLog.SearchTimeVariable(refVar.Epoch) //Get what this time is by variable
	if tvErr != nil {
		return 999999999999999999, tvErr
	}

	tvEpoch, errTvEpoch := tv.SolveEpoch(refVar.BootNumber, refVar.Uptime)
	if errTvEpoch != nil {
		return NsEpoch(0), errTvEpoch
	}

	diff := tvEpoch.Diff(refVar.Epoch)
	if diff < 0 {
		diff = -diff
	}

	return diff, nil
}
