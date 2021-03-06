package dataunlynx_test

import (
	"github.com/dedis/onet/log"
	"github.com/lca1/unlynx/lib"
	"github.com/lca1/unlynx/services/default/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

const filename = "unlynx_test_data.txt"
const numDPs = 2
const numEntries = 10
const numEntriesFiltered = 5
const numGroupsClear = 0
const numGroupsEnc = 2
const numWhereClear = 0
const numWhereEnc = 2
const numAggrClear = 0
const numAggrEnc = 2

var num_type = [...]int64{2, 5}

var test_data map[string][]libunlynx.DpClearResponse

func TestAllPossibleGroups(t *testing.T) {
	dataunlynx.Groups = make([][]int64, 0)

	group := make([]int64, 0)
	dataunlynx.AllPossibleGroups(num_type[:], group, 0)

	num_elem := 1
	for _, el := range num_type {
		num_elem = num_elem * int(el)
	}
	assert.Equal(t, num_elem, len(dataunlynx.Groups), "Some elements are missing")
}

func TestGenerateData(t *testing.T) {
	test_data = dataunlynx.GenerateData(numDPs, numEntries, numEntriesFiltered, numGroupsClear, numGroupsEnc,
		numWhereClear, numWhereEnc, numAggrClear, numAggrEnc, num_type[:], true)
}

func TestWriteDataToFile(t *testing.T) {
	dataunlynx.WriteDataToFile(filename, test_data)
}

func TestReadDataFromFile(t *testing.T) {
	dataunlynx.ReadDataFromFile(filename)
}

func TestCompareClearResponses(t *testing.T) {
	dataunlynx.ReadDataFromFile(filename)
	assert.Equal(t, test_data, dataunlynx.ReadDataFromFile(filename), "Data should be the same")
}

func TestComputeExpectedResult(t *testing.T) {
	log.Lvl1(dataunlynx.ComputeExpectedResult(dataunlynx.ReadDataFromFile(filename), 1, false))
	assert.Equal(t, dataunlynx.CompareClearResponses(dataunlynx.ComputeExpectedResult(test_data, 1, false), dataunlynx.ComputeExpectedResult(dataunlynx.ReadDataFromFile(filename), 1, false)), true, "Result should be the same")
	assert.Equal(t, dataunlynx.CompareClearResponses(dataunlynx.ComputeExpectedResult(test_data, 1, true), dataunlynx.ComputeExpectedResult(dataunlynx.ReadDataFromFile(filename), 1, true)), true, "Result should be the same")
}
