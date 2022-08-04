/*

Struct TimeFileDb is conversion layer for storing TimeVariables to disk in reliable way
Current implementation persist data on disk but keeps content cached in mem for fast search.

This might be optimized later
*/

package timegopher

import (
	"fmt"

	"github.com/hjkoskel/fixregsto"
)

type TimeFileDb struct {
	sto      fixregsto.FixRegSto //Store and restore here
	storeRTC bool                //false= only boot and uptime
	mem      TimeVariableList    //Primary place to keep values
}

//CreateTimeFileDb restores content from FixRegSto storage and initializes TimeFileDb struct
func CreateTimeFileDb(storage fixregsto.FixRegSto, storeRTC bool) (TimeFileDb, error) {
	raw, readErr := storage.ReadAll()
	if readErr != nil {
		return TimeFileDb{}, fmt.Errorf("error on ReadAll on CreateTimeFileDb err=%v", readErr.Error())
	}
	mem, errParse := ParseTimeVariableList(raw, storeRTC)
	return TimeFileDb{sto: storage, storeRTC: storeRTC, mem: mem}, errParse
}

func (p *TimeFileDb) Insert(t TimeVariable) error { //INSERT only cumulative values
	binarr, errbin := t.ToBinary(p.storeRTC)
	if errbin != nil {
		return fmt.Errorf("Insert error, binary coding %#v failed %v", t, errbin)
	}
	n := p.mem.Len()
	if 0 < n { //If there are points, check that new variable is t is really after. Not before or same
		if !p.mem[p.mem.Len()-1].Before(t) {
			return fmt.Errorf("inserted time t=%#v is before latest entry %#v", t, p.mem[p.mem.Len()-1])
		}
	}
	_, errWrite := p.sto.Write(binarr)
	if errWrite != nil {
		return errWrite
	}
	p.mem = append(p.mem, t)

	return nil
}

func (p *TimeFileDb) GetLatestN(n int) ([]TimeVariable, error) {
	maxN := p.mem.Len()
	if maxN < n {
		return p.mem, nil
	}
	return p.mem[maxN-n : p.mem.Len()], nil

}
func (p *TimeFileDb) GetOnBoot(boot int32) ([]TimeVariable, error) {
	result := []TimeVariable{}
	for _, v := range p.mem {
		if v.BootNumber == boot {
			result = append(result, v)
		}
	}
	return result, nil

}
func (p *TimeFileDb) GetFirstN(n int) ([]TimeVariable, error) {
	if p.mem.Len() < n {
		return p.mem, nil
	}
	return p.mem[0:n], nil
}

func (p *TimeFileDb) All() ([]TimeVariable, error) {
	return p.mem, nil
}

func (p *TimeFileDb) Len() (int, error) {
	return p.mem.Len(), nil
}

func (p *TimeFileDb) SolveEpoch(boot int32, uptime NsUptime) (NsEpoch, error) {
	return p.mem.SolveEpoch(boot, uptime) //TODO optimize? No need implementation in arr. Search boot and search time
}

//Search vs solve
func (p *TimeFileDb) SolveBootNumber(epoch NsEpoch) (int32, error) {
	return p.mem.SolveBootNumber(epoch)
}

func (p *TimeFileDb) SearchTimeVariable(epoch NsEpoch) (TimeVariable, error) {
	return p.mem.SearchTimeVariable(epoch)
}
