package libunlynxproofs_test

import (
	"github.com/lca1/unlynx/lib"
	"github.com/lca1/unlynx/lib/proofs"
	"github.com/lca1/unlynx/lib/tools"
	"github.com/lca1/unlynx/protocols"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAggregationProof(t *testing.T) {
	tab1 := []int64{1, 2, 3, 6}
	testCipherVect1 := *libunlynx.EncryptIntVector(pubKey, tab1)

	tab2 := []int64{2, 4, 8, 6}
	testCipherVect2 := *libunlynx.EncryptIntVector(pubKey, tab2)

	det1 := testCipherVect2
	det2 := testCipherVect1
	det3 := testCipherVect2

	protocolsunlynx.TaggingDet(&det1, secKey, secKey, pubKey, true)
	deterministicGroupAttributes := make(libunlynx.DeterministCipherVector, len(det1))
	for j, c := range det1 {
		deterministicGroupAttributes[j] = libunlynx.DeterministCipherText{Point: c.C}
	}
	newDetResponse1 := libunlynx.FilteredResponseDet{Fr: libunlynx.FilteredResponse{GroupByEnc: testCipherVect2, AggregatingAttributes: testCipherVect1}, DetTagGroupBy: deterministicGroupAttributes.Key()}

	protocolsunlynx.TaggingDet(&det2, secKey, secKey, pubKey, true)

	deterministicGroupAttributes = make(libunlynx.DeterministCipherVector, len(det2))
	for j, c := range det2 {
		deterministicGroupAttributes[j] = libunlynx.DeterministCipherText{Point: c.C}
	}
	newDetResponse2 := libunlynx.FilteredResponseDet{Fr: libunlynx.FilteredResponse{GroupByEnc: testCipherVect1, AggregatingAttributes: testCipherVect1}, DetTagGroupBy: deterministicGroupAttributes.Key()}

	protocolsunlynx.TaggingDet(&det3, secKey, secKey, pubKey, true)
	deterministicGroupAttributes = make(libunlynx.DeterministCipherVector, len(det3))
	for j, c := range det3 {
		deterministicGroupAttributes[j] = libunlynx.DeterministCipherText{Point: c.C}
	}
	newDetResponse3 := libunlynx.FilteredResponseDet{Fr: libunlynx.FilteredResponse{GroupByEnc: testCipherVect2, AggregatingAttributes: testCipherVect1}, DetTagGroupBy: deterministicGroupAttributes.Key()}

	detResponses := make([]libunlynx.FilteredResponseDet, 3)
	detResponses[0] = newDetResponse1
	detResponses[1] = newDetResponse2
	detResponses[2] = newDetResponse3

	comparisonMap := make(map[libunlynx.GroupingKey]libunlynx.FilteredResponse)
	for _, v := range detResponses {
		libunlynxtools.AddInMap(comparisonMap, v.DetTagGroupBy, v.Fr)
	}

	PublishedAggregationProof := libunlynxproofs.AggregationProofCreation(detResponses, comparisonMap)
	assert.True(t, libunlynxproofs.AggregationProofVerification(PublishedAggregationProof))

	detResponses[0] = detResponses[1]
	PublishedAggregationProof = libunlynxproofs.AggregationProofCreation(detResponses, comparisonMap)
	assert.False(t, libunlynxproofs.AggregationProofVerification(PublishedAggregationProof))
}
