/*
Example use of timegopher

TODO log?
/sys/class/thermal/thermal_zone0/temp
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
	"timegopher"
)

var rtcSystem timegopher.TimeGopher

func dumbDbToFile(db *timegopher.TimeFileDb, filename string) error {
	_, errCount := (*db).Len()
	if errCount != nil {
		return errCount
	}
	arr, arrErr := (*db).GetFirstN(int(999999999))
	if arrErr != nil {
		return arrErr
	}
	return ioutil.WriteFile(filename, []byte(timegopher.TimeVariableList(arr).String()), 0666)
}

func debugDumpFilebases() error {
	err := os.MkdirAll("./rtcDebugDump/", 0755)
	if err != nil {
		return err
	}
	fmt.Printf("\n\n!!!DUMP UNCERTAIN!!!\n")
	err = dumbDbToFile(rtcSystem.UncertainRtcSyncLog, "./rtcDebugDump/UncertainRtcSyncLog.csv")
	if err != nil {
		return err
	}
	fmt.Printf("\n\n!!!RTCSYNCLOG\n")
	err = dumbDbToFile(rtcSystem.RtcSyncLog, "./rtcDebugDump/RtcSyncLog.csv")
	if err != nil {
		return err
	}
	fmt.Printf("\n\n!!!STARTLOG\n")
	err = dumbDbToFile(rtcSystem.StartLog, "./rtcDebugDump/StartLog.csv")
	if err != nil {
		return err
	}
	fmt.Printf("\n\n!!!STOPLOG\n")
	err = dumbDbToFile(rtcSystem.StopLog, "./rtcDebugDump/StopLog.csv")
	if err != nil {
		return err
	}
	fmt.Printf("\n\n!!!LASTLOG\n")
	err = dumbDbToFile(rtcSystem.LastLog, "./rtcDebugDump/LastLog.csv")
	if err != nil {
		return err
	}

	return nil
}

func readExampleTemperature() (float64, error) {
	byt, readErr := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if readErr != nil {
		return 0, readErr
	}
	i, parseErr := strconv.ParseInt(strings.TrimSpace(string(byt)), 10, 64)
	if parseErr != nil {
		return 0, parseErr
	}
	return float64(i) / 1000, nil
}

const (
	MAXWORKFILEROWS = 256
)

const DATALIB = "./rtcdata"
const WORKCSV = "work.csv" //Just for demo.. appending

func dataloggerRoutine(ch chan DemoDataPoint) error {
	//restore worklog from disk
	workdata, errRestore := LoadDemoContent(WORKCSV)
	if errRestore != nil {
		fmt.Printf("Was not able to restore demo work content (%s), continue err=%v\n", WORKCSV, errRestore.Error())
	}

	//If in sync, then throw worklog to export log
	for {
		//Get measurements
		newPoint := <-ch

		workdata = append(workdata, newPoint)
		errSave := workdata.SaveDemoContent(WORKCSV)
		if errSave != nil {
			return fmt.Errorf("Was not able to save example data %v\n", errSave.Error())
		}

		//Append those, sync to disk
		time.Sleep(time.Millisecond * 100)
	}
}

func main() {
	fmt.Printf("--RTC example--\n")
	os.MkdirAll(DATALIB, 0755)
	var createErr error

	/*
		isFirstRun, errfirstrun := timegopher.FirstCallAfterBoot("/tmp/warmstart")
		if errfirstrun != nil {
			fmt.Printf("errfirstrun %#v\n", errfirstrun.Error())
		}*/

	rtcSystem, createErr = timegopher.CreateDefaultTimeGopher(DATALIB, timegopher.TimeVariable{})
	if createErr != nil {
		fmt.Printf("CreateErr %v\n", createErr)
		return
	}

	dumpErr := debugDumpFilebases()
	if dumpErr != nil {
		fmt.Printf("DUMP fail %v\n", dumpErr.Error())
		return
	}

	//fmt.Printf("--- STARTUP LOG ---\n%s\n\n", rtcSystem.StartupLog.GetLatestN(10))

	firstrun := rtcSystem.IsColdStart()
	if firstrun {
		fmt.Printf("THIS IS FIRST RUN AFTER BOOT!!!\n")
	}

	latest, errLatest := rtcSystem.GetLatestTime()
	if errLatest != nil {
		fmt.Printf("Latest time fail %v\n", errLatest.Error())
		return
	}
	fmt.Printf("Now the Latest %#v\n", latest)

	measurementResults := make(chan DemoDataPoint)

	go func() {
		var synced bool
		var errRtc error
		for {
			if !synced {
				synced, errRtc = timegopher.RtcIsSynced_adjtimex()
				if errRtc != nil {
					fmt.Printf("errRTC=%v\n", errRtc)
					return
				}
				fmt.Printf("synced=%v\n", synced)
			}
			errRefresh := rtcSystem.Refresh(time.Now(), synced)
			if errRefresh != nil {
				fmt.Printf("ERROR IN REFRESH %v (should not happen) TODO HANDLE\n", errRefresh.Error())
				os.Exit(-1)
			}
			time.Sleep(1000 * time.Millisecond)
			n, _ := rtcSystem.RtcSyncLog.Len()
			fmt.Printf("RTC sync log is now %v\n", n)

		}
	}()

	//Goroutine for writing log
	go dataloggerRoutine(measurementResults)

	for {
		temperature, errTemperature := readExampleTemperature()
		if errTemperature != nil {
			fmt.Printf("TEMP fail %v\n", errTemperature)
		}
		//fmt.Printf("T=%.1f\n", temperature)
		tv, tvErr := rtcSystem.Convert(time.Now())
		if tvErr != nil {
			fmt.Printf("convert time fail %v\n", tvErr)
			return
		}

		m := DemoDataPoint{
			Temperature: temperature,
			Timestamp:   tv,
		}

		measurementResults <- m
		time.Sleep(1000 * time.Millisecond)
	}
}
