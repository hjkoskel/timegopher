/*
Demo content

Routines for loading and saving simple time series.
Only for example. Bad and unrealiable
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"timegopher"
)

type DemoDataPoint struct {
	Temperature float64
	Timestamp   timegopher.TimeVariable
}

func (a DemoDataPoint) String() string {
	t := time.Unix(0, int64(a.Timestamp.Epoch))
	return fmt.Sprintf("%.1f\t%v\t%v\t%v\t%s",
		a.Temperature,
		a.Timestamp.BootNumber,
		a.Timestamp.Epoch,
		a.Timestamp.Uptime,
		t)
}

func ParseDemoDataPoint(s string) (DemoDataPoint, error) {
	result := DemoDataPoint{}

	nFields, errScanf := fmt.Sscanf(s, "%f\t%v\t%v\t%v",
		&result.Temperature,
		&result.Timestamp.BootNumber,
		&result.Timestamp.Epoch,
		&result.Timestamp.Uptime)
	if errScanf != nil {
		return result, fmt.Errorf("Error parsing row %v err=%v", s, errScanf.Error())
	}
	if nFields != 4 {
		return result, fmt.Errorf("INVALID FORMAT AT ROW %v", s)
	}
	return result, nil
}

type DemoContent []DemoDataPoint

func LoadDemoContent(filename string) (DemoContent, error) {
	result := []DemoDataPoint{}
	byt, err := os.ReadFile(filename)
	if err != nil {
		return DemoContent{}, err
	}
	rows := strings.Split(string(byt), "\n")
	for rownumber, row := range rows {
		s := strings.TrimSpace(row)
		if len(s) == 0 {
			continue
		}
		v, parseErr := ParseDemoDataPoint(s)
		if parseErr != nil {
			return result, fmt.Errorf("Parse error %v at row %v", parseErr, rownumber)
		}

		result = append(result, v)
	}
	return result, nil
}

func (p *DemoContent) SaveDemoContent(filename string) error {
	var sb strings.Builder
	for _, v := range *p {
		sb.WriteString(v.String() + "\n")
	}
	return ioutil.WriteFile(filename, []byte(sb.String()), 0666)
}
