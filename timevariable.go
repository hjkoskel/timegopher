/*
Time variable.

Intelligent way to store timestamp even with temporary uncertain RTC.
Allows resolving epoch later based on known TimeVariables
*/

package timegopher

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type NsEpoch int64
type NsUptime int64

//Used for sanity check, was a joke at first
const EPOCH70S = 10 * 365 * 24 * 60 * 60 * 1000 * 1000 * 1000

//TimeVariable is way to store timestamps. Epoch can solved by know TimeVariables with same bootNumber
type TimeVariable struct {
	BootNumber int32
	Uptime     NsUptime
	Epoch      NsEpoch
}

//Diff compares NsEpoch, returns always positive
func (p NsEpoch) Diff(ref NsEpoch) NsEpoch {
	if p < ref {
		return ref - p
	}
	return p - ref
}

//Diff compares uptime, returns always positive
func (p NsUptime) Diff(ref NsUptime) NsUptime {
	if p < ref {
		return ref - p
	}
	return p - ref
}

//Seconds converts to NsEpoch
func (p NsEpoch) Seconds() float64 { //For debug purposes, conversions
	return float64(p) / float64(1000*1000*1000)
}

//ToBinary creates binary presentation of time variable. Some variables do not need epoch
func (p *TimeVariable) ToBinary(storeRTC bool) ([]byte, error) {
	buf := new(bytes.Buffer) //Without RTC (32+64)/8=12, with RTC 20
	err := binary.Write(buf, binary.LittleEndian, p.BootNumber)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, p.Uptime)
	if err != nil {
		return nil, err
	}
	if p.Uptime <= 0 {
		return nil, fmt.Errorf("ToBinary: Uptime is %v", p.Uptime)
	}

	if storeRTC {
		if p.Epoch < EPOCH70S {
			return nil, fmt.Errorf("ToBinary: Missing epoch, 1970's not supported")
		}

		err = binary.Write(buf, binary.LittleEndian, p.Epoch)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

//ParseTimeVariable parses TimeVariable from binary format
func ParseTimeVariable(raw []byte, storeRTC bool) (TimeVariable, error) {
	if storeRTC {
		if len(raw) != 20 {
			return TimeVariable{}, fmt.Errorf("invalid size %v for timevariable with RTC", len(raw))
		}
		result := TimeVariable{
			BootNumber: int32(binary.LittleEndian.Uint32(raw[0:4])),
			Uptime:     NsUptime(binary.LittleEndian.Uint64(raw[4:12])),
			Epoch:      NsEpoch(binary.LittleEndian.Uint64(raw[12:20])),
		}
		if result.Epoch < EPOCH70S { //Catch errors early but return result still
			return result, fmt.Errorf("ParseTimeVariable:Missing epoch, 1970's not supported")
		}
		return result, nil
	}
	//without RTC readings
	if len(raw) != 12 {
		return TimeVariable{}, fmt.Errorf("invalid size %v for timevariable without RTC", len(raw))
	}
	result := TimeVariable{
		BootNumber: int32(binary.LittleEndian.Uint32(raw[0:4])),
		Uptime:     NsUptime(binary.LittleEndian.Uint64(raw[4:12])),
	}
	if result.Uptime <= 0 {
		return result, fmt.Errorf("parseTimeVariable: Uptime is %v", result.Uptime)
	}

	return result, nil
}

//SolveEpoch calculates propotional epoch from uptime
func (p *TimeVariable) SolveEpoch(bootNumber int32, uptime NsUptime) (NsEpoch, error) {
	if p.BootNumber != bootNumber {
		return 0, fmt.Errorf("not same boot number %v (must be %v)", p.BootNumber, bootNumber)
	}
	if p.Epoch < EPOCH70S {
		return 0, fmt.Errorf("SolveEpoch:Missing epoch, 1970's not supported")
	}
	if p.BootNumber < 0 {
		return 0, fmt.Errorf("invalid boot number %v", p.BootNumber)
	}
	if uptime <= 0 {
		return 0, fmt.Errorf("SolveEpoch: Uptime is invalid %v", uptime)
	}

	bootEpoch := p.Epoch - NsEpoch(p.Uptime)
	return bootEpoch + NsEpoch(uptime), nil
}

//SolveUptime calculates propotional uptime from epoch and does validity check
func (p *TimeVariable) SolveUptime(bootNumber int32, epoch NsEpoch) (NsUptime, error) {
	if p.BootNumber != bootNumber {
		return 0, fmt.Errorf("not same boot number %v (must be %v)", p.BootNumber, bootNumber)
	}
	if p.BootNumber < 0 {
		return 0, fmt.Errorf("invalid boot number %v", p.BootNumber)
	}
	if p.Epoch < EPOCH70S {
		return 0, fmt.Errorf("SolveUptime:Missing epoch, 1970's not supported")
	}
	return p.Uptime + NsUptime(epoch-p.Epoch), nil
}

//Equal reports whether p and u are equal
func (p *TimeVariable) Equal(u TimeVariable) bool {
	return p.BootNumber == u.BootNumber && p.Uptime == u.Uptime && p.Epoch == u.Epoch
}

//After reports whether the TimeVariable instant p is after u.
func (p *TimeVariable) After(u TimeVariable) bool {
	//If uptime is available then use that
	if 0 < p.Uptime && 0 < u.Uptime {
		if u.BootNumber == p.BootNumber {
			//Same boot, easy
			return u.Uptime < p.Uptime
		}
		return u.BootNumber < p.BootNumber
	}
	return u.Epoch < p.Epoch
}

//After reports whether the TimeVariable instant p is before u.
func (p *TimeVariable) Before(u TimeVariable) bool {
	if p.Equal(u) {
		return false //Equal is not before
	}
	return !p.After(u)
}
