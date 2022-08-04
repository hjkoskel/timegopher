package timegopher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const TESTEPOCH0 = 1658334406982 * 1000 * 1000

func TestMiscTimeVariable(t *testing.T) {
	dut := TimeVariable{BootNumber: 3, Epoch: TESTEPOCH0, Uptime: 5000}
	dutSame := TimeVariable{BootNumber: 3, Epoch: TESTEPOCH0, Uptime: 5000}
	dutAnotherBoot := TimeVariable{BootNumber: 4, Epoch: TESTEPOCH0 + 100000, Uptime: 5000}

	assert.Equal(t, dut.Equal(dutSame), true)
	assert.Equal(t, dut.Equal(dutAnotherBoot), false)

}

func TestBin(t *testing.T) {
	dut := TimeVariable{BootNumber: 3, Epoch: TESTEPOCH0, Uptime: 5000}

	testbinWithRTC, errBinWithRTC := dut.ToBinary(true)
	assert.Equal(t, nil, errBinWithRTC)

	dutRefWithRTC, dutRefWithRTCErr := ParseTimeVariable(testbinWithRTC, true)
	assert.Equal(t, nil, dutRefWithRTCErr)
	assert.Equal(t, dutRefWithRTC, dut)

	testbin, errBin := dut.ToBinary(false)
	assert.Equal(t, nil, errBin)

	dutRef, dutRefErr := ParseTimeVariable(testbin, false)
	assert.Equal(t, nil, dutRefErr)
	assert.Equal(t, dutRef, TimeVariable{BootNumber: 3, Epoch: 0, Uptime: 5000})

	//Test errors
	_, parsefail1 := ParseTimeVariable(testbinWithRTC, false)
	_, parsefail2 := ParseTimeVariable(testbin, true)
	assert.NotEqual(t, nil, parsefail1)
	assert.NotEqual(t, nil, parsefail2)

}

func TestCalc(t *testing.T) {
	dut := TimeVariable{BootNumber: 3, Epoch: TESTEPOCH0, Uptime: 5000}
	epo1, epo1err := dut.SolveEpoch(3, dut.Uptime+1000)
	assert.Equal(t, dut.Epoch+1000, epo1)
	assert.Equal(t, nil, epo1err)

	_, failepo := dut.SolveEpoch(2, dut.Uptime+1000)
	assert.NotEqual(t, nil, failepo)

	ut1, ut1err := dut.SolveUptime(3, epo1)
	assert.Equal(t, dut.Uptime+1000, ut1)
	assert.Equal(t, nil, ut1err)

	_, failut := dut.SolveUptime(2, epo1)
	assert.NotEqual(t, nil, failut)

	assert.Equal(t, false, dut.After(TimeVariable{BootNumber: 3, Uptime: 6000}))
	assert.Equal(t, true, dut.After(TimeVariable{BootNumber: 3, Uptime: 4000}))
	assert.Equal(t, true, dut.After(TimeVariable{BootNumber: 2, Uptime: 9999}))
	assert.Equal(t, false, dut.After(TimeVariable{BootNumber: 4, Uptime: 10}))

	assert.Equal(t, true, dut.Before(TimeVariable{BootNumber: 3, Uptime: 6000}))
	assert.Equal(t, false, dut.Before(TimeVariable{BootNumber: 3, Uptime: 4000}))
	assert.Equal(t, false, dut.Before(TimeVariable{BootNumber: 2, Uptime: 9999}))
	assert.Equal(t, true, dut.Before(TimeVariable{BootNumber: 4, Uptime: 10}))

	onlyepoch := TimeVariable{Epoch: 10}
	assert.Equal(t, false, onlyepoch.After(TimeVariable{Epoch: 11}))
	assert.Equal(t, false, onlyepoch.After(TimeVariable{Epoch: 10})) //Keep equal same
	assert.Equal(t, true, onlyepoch.After(TimeVariable{Epoch: 9}))

	assert.Equal(t, true, onlyepoch.Before(TimeVariable{Epoch: 11}))
	assert.Equal(t, false, onlyepoch.Before(TimeVariable{Epoch: 10})) //Keep equal same
	assert.Equal(t, false, onlyepoch.Before(TimeVariable{Epoch: 9}))

}
