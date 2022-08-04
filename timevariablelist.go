/*
Time variable list
Aluksi vain array ja ehkä muistinvaraisena ok

Myöhemmin levytallennusta varten Deltakoodaten ja muutenkin että mahd vähän vaihtelua että pakkaantuu

- Alkuun eka sellaisenaan
- Bootnumber, vaan delta
- Epoch koodataan aina boottimeksi, liki sama saman bootin entryissä


*/

package timegopher

import (
	"fmt"
	"strings"
)

type TimeVariableList []TimeVariable

//Len number of TimeVariables in list
func (e TimeVariableList) Len() int {
	return len(e)
}

//Less function for sorting
func (e TimeVariableList) Less(i, j int) bool {
	if e[i].BootNumber == e[j].BootNumber {
		return e[i].Uptime < e[j].Uptime
	}
	return e[i].BootNumber < e[j].BootNumber
}

//Swap function for sorting
func (e TimeVariableList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

//SolveEpoch picks TimeVariable entry at defined boot number just before or at uptime and uses that for solving epoch
func (p *TimeVariableList) SolveEpoch(bootNumber int32, uptime NsUptime) (NsEpoch, error) {
	if len(*p) == 0 {
		return 0, fmt.Errorf("no data in TimeVariableList, solving bootNumber=%v, uptime=%v", bootNumber, uptime)
	}
	//Get "just before" point
	index := -1
	for i, v := range *p { //TODO optimized search later
		if bootNumber < v.BootNumber {
			return 0, fmt.Errorf("not found point for boot %v", bootNumber)
		}
		if v.BootNumber == bootNumber {
			if v.Uptime <= uptime {
				index = i
			}
		}
	}

	if index < 0 {
		if (*p)[p.Len()-1].BootNumber == bootNumber {
			return (*p)[p.Len()-1].SolveEpoch(bootNumber, uptime)
		}
		return 0, fmt.Errorf("points not found boot %v", bootNumber)
	}
	return (*p)[index].SolveEpoch(bootNumber, uptime)
}

//SolveBootNumber searches from list. Assumption is that array is sorted.. old at low indexes... newest at higher indexes
func (p *TimeVariableList) SolveBootNumber(epoch NsEpoch) (int32, error) {
	if len(*p) == 0 {
		return -1, fmt.Errorf("no data, while solving boot number from epoch %v", epoch)
	}
	result := int32(-1)
	for _, v := range *p {
		bootTime := v.Epoch - NsEpoch(v.Uptime)
		if (bootTime <= epoch) && (bootTime != 0) {
			result = v.BootNumber //Last boot stays as result
		}
	}
	if result == -1 {
		return -1, fmt.Errorf("was not able find epoch before %v", epoch)
	}
	return result, nil
}

//SearchTimeVariable from list that is nearest to parameter epoch (at same boot)
func (p *TimeVariableList) SearchTimeVariable(epoch NsEpoch) (TimeVariable, error) {
	bootNumber, errBootNumber := p.SolveBootNumber(epoch)
	if errBootNumber != nil {
		return TimeVariable{}, errBootNumber
	}

	dataInBoot := p.GetVariablesInBoot(bootNumber)
	if len(dataInBoot) == 0 { //This can not really happen but if happens then it is bug or some internal error
		return TimeVariable{}, fmt.Errorf("internal error")
	}
	result := dataInBoot[0] //Initial value
	diff := result.Epoch.Diff(epoch)
	for _, v := range *p {
		d := v.Epoch.Diff(epoch)
		if d < diff {
			result = v
			diff = d
		}
	}
	return result, nil
}

//String representation from TimeVariableList with newline at end
func (p TimeVariableList) String() string {
	var sb strings.Builder
	for _, a := range p {
		if a.Epoch == 0 { //Only if not synced. Errors are handled before
			sb.WriteString(fmt.Sprintf("%v\t%v\n", a.BootNumber, a.Uptime))
		} else {
			sb.WriteString(fmt.Sprintf("%v\t%v\t%v\n", a.BootNumber, a.Uptime, a.Epoch))
		}
	}
	return sb.String()
}

//GetFirstsInBoot picks first entries from each boot by index
func (p *TimeVariableList) GetFirstsInBoot() TimeVariableList {
	result := []TimeVariable{}
	bn := int32(-100) //TODO ERR
	for _, a := range *p {
		if a.BootNumber != bn {
			result = append(result, a)
			bn = a.BootNumber
		}
	}
	return result
}

//GetVariablesInBoot give all entries with specific boot number
func (p *TimeVariableList) GetVariablesInBoot(boot int32) TimeVariableList {
	result := []TimeVariable{}
	for _, a := range *p {
		if a.BootNumber == boot {
			result = append(result, a)
		}
	}
	return result
}

//ParseTimeVariableList parses from raw byte array. Check length validity
func ParseTimeVariableList(raw []byte, storeRTC bool) (TimeVariableList, error) {
	size := 12
	if storeRTC {
		size = 20
	}

	if len(raw)%size != 0 {
		return []TimeVariable{}, fmt.Errorf("must be multiple of %v (len=%v)", size, raw)
	}

	result := make([]TimeVariable, len(raw)/size)
	for i := range result {
		var errParse error
		arr := raw[i*size : (i+1)*size]
		result[i], errParse = ParseTimeVariable(arr, storeRTC)
		if errParse != nil {
			return result, errParse
		}
	}
	return result, nil
}
