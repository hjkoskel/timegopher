package timegopher

import (
	"testing"

	"github.com/hjkoskel/fixregsto"

	"github.com/stretchr/testify/assert"
)

func TestFilebase(t *testing.T) {
	memconf := fixregsto.MemloopConf{
		RecordSize: RECORDSIZE_TIMEVARIABLE_RTC,
		MaxRecords: 8,
	}

	mem, memCreateErr := memconf.InitMemLoop()
	if memCreateErr != nil {
		t.Error(memCreateErr)
	}

	dut, errCreate := CreateTimeFileDb(&mem, true)
	if errCreate != nil {
		t.Error(errCreate)
	}
	insertErr := dut.Insert(TimeVariable{BootNumber: 1, Uptime: 200, Epoch: TESTEPOCH0 + 1000}) //boot at 1000-200=800
	if insertErr != nil {
		t.Error(insertErr)
	}
	insertErr = dut.Insert(TimeVariable{BootNumber: 3, Uptime: 1000, Epoch: TESTEPOCH0 + 10000}) //boot at 9000
	if insertErr != nil {
		t.Error(insertErr)
	}
	insertErr = dut.Insert(TimeVariable{BootNumber: 3, Uptime: 1100, Epoch: TESTEPOCH0 + 10100}) //boot at 9000
	if insertErr != nil {
		t.Error(insertErr)
	}

	n, lenErr := dut.Len()
	if lenErr != nil {
		t.Error(lenErr)
	}
	assert.Equal(t, 3, n)

	boot3, boot3err := dut.GetOnBoot(3)
	if boot3err != nil {
		t.Error(boot3)
	}
	assert.Equal(t, len(boot3), 2)

	latest2, latest2err := dut.GetLatestN(2)
	assert.Equal(t, []TimeVariable{
		TimeVariable{BootNumber: 3, Uptime: 1000, Epoch: TESTEPOCH0 + 10000},
		TimeVariable{BootNumber: 3, Uptime: 1100, Epoch: TESTEPOCH0 + 10100},
	}, latest2)
	assert.Equal(t, nil, latest2err)

	testEpo, testEpoErr := dut.SolveEpoch(3, 1200)
	assert.Equal(t, nil, testEpoErr)
	assert.Equal(t, NsEpoch(TESTEPOCH0+10200), testEpo)

	_, bootNofounderr := dut.SolveBootNumber(TESTEPOCH0 + 500)
	assert.NotNil(t, bootNofounderr)

	bootIs1, bootIs1err := dut.SolveBootNumber(TESTEPOCH0 + 8999)
	assert.Equal(t, nil, bootIs1err)
	assert.Equal(t, int32(1), bootIs1)

	bootIs3, bootIs3err := dut.SolveBootNumber(TESTEPOCH0 + 10000)
	assert.Equal(t, nil, bootIs3err)
	assert.Equal(t, int32(3), bootIs3)

	searched1, errSearched1 := dut.SearchTimeVariable(TESTEPOCH0 + 801)
	assert.Equal(t, nil, errSearched1)
	assert.Equal(t, TimeVariable{BootNumber: 1, Uptime: 200, Epoch: TESTEPOCH0 + 1000}, searched1)

	searched2, errSearched2 := dut.SearchTimeVariable(TESTEPOCH0 + 10005)
	assert.Equal(t, nil, errSearched2)
	assert.Equal(t, TimeVariable{BootNumber: 3, Uptime: 1000, Epoch: TESTEPOCH0 + 10000}, searched2)

}
