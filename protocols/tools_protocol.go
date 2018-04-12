//Go file that regroup different tools functions used in the protocols
package protocolsunlynx

import (
    "github.com/lca1/unlynx/lib"
)

// CypherTextArray build from a ProcessResponse array
func PRToCipherTextArray(p []libunlynx.ProcessResponse) []libunlynx.CipherText {
    cipherTexts := make([]libunlynx.CipherText, 0)

    for _, v := range p {
        cipherTexts = append(cipherTexts, v.WhereEnc...)
        cipherTexts = append(cipherTexts, v.GroupByEnc...)
    }

    return cipherTexts
}

// ProcessResponseDet build from DeterministicCipherVector and the ProcessResponse
func DCVToProcessResponseDet(detCt libunlynx.DeterministCipherVector,
    targetOfSwitch []libunlynx.ProcessResponse) []libunlynx.ProcessResponseDet {
    result := make([]libunlynx.ProcessResponseDet, len(targetOfSwitch))

    pos := 0
    for i := range result {
        whereEncLen := len(targetOfSwitch[i].WhereEnc)
        deterministicWhereAttributes := make([]libunlynx.GroupingKey, whereEncLen)
        for j, c := range detCt[pos : pos+whereEncLen] {
            deterministicWhereAttributes[j] = libunlynx.GroupingKey(c.String())
        }
        pos += whereEncLen

        groupByLen := len(targetOfSwitch[i].GroupByEnc)
        deterministicGroupAttributes := make(libunlynx.DeterministCipherVector, groupByLen)
        copy(deterministicGroupAttributes, detCt[pos : pos+groupByLen])
        pos += groupByLen

        result[i] = libunlynx.ProcessResponseDet{PR: targetOfSwitch[i], DetTagGroupBy: deterministicGroupAttributes.Key(),
            DetTagWhere: deterministicWhereAttributes}
    }

    return result
}

// FilterResponse transformed into a CipherVector. Return also the lengths necessary to rebuild the function
func FilteredResponseToCipherVector(fr []libunlynx.FilteredResponse) (libunlynx.CipherVector, [][]int) {
    cv := make(libunlynx.CipherVector, 0)
    lengths := make([][]int, len(fr))

    for i, v := range fr {
        lengths[i] = make([]int, 2)
        cv = append(cv, v.GroupByEnc...)
        lengths[i][0] = len(v.GroupByEnc)
        cv = append(cv, v.AggregatingAttributes...)
        lengths[i][1] = len(v.AggregatingAttributes)
    }

    return cv, lengths
}

//CipherVector rebuild a FilteredResponse with the length of the FilteredResponse given in FilteredResponseToCipherVector
func CipherVectorToFilteredResponse(cv libunlynx.CipherVector, lengths [][]int) []libunlynx.FilteredResponse {
    filteredResponse := make([]libunlynx.FilteredResponse, len(lengths))

    pos := 0
    for i, length := range lengths {
        filteredResponse[i].GroupByEnc = make(libunlynx.CipherVector, length[0])
        copy(filteredResponse[i].GroupByEnc, cv[pos:pos+length[0]])
        pos += length[0]

        filteredResponse[i].AggregatingAttributes = make(libunlynx.CipherVector, length[1])
        copy(filteredResponse[i].AggregatingAttributes, cv[pos:pos+length[1]])
        pos += length[1]
    }

    return filteredResponse
}
