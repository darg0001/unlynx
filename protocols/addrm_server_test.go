package protocolsunlynx_test

import (
	"github.com/dedis/kyber/util/random"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/dedis/onet/network"
	"github.com/lca1/unlynx/lib"
	"github.com/lca1/unlynx/protocols"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddRmServer(t *testing.T) {
	local := onet.NewLocalTest(libunlynx.SuiTe)
	_, _, tree := local.GenTree(1, true)

	defer local.CloseAll()

	rootInstance, err := local.CreateProtocol("AddRmServer", tree)
	if err != nil {
		t.Fatal("Couldn't start protocol:", err)
	}
	protocol := rootInstance.(*protocolsunlynx.AddRmServerProtocol)

	secKey := libunlynx.SuiTe.Scalar().Pick(random.New())
	pubKey := libunlynx.SuiTe.Point().Mul(secKey, libunlynx.SuiTe.Point().Base())

	secKeyAddRm := libunlynx.SuiTe.Scalar().Pick(random.New())

	//substraction
	secKeyAfter := libunlynx.SuiTe.Scalar().Sub(secKey, secKeyAddRm)

	tab := []int64{10, 10}

	expectedResults := make([]int64, len(tab))
	expectedResults[0] = 10
	expectedResults[1] = 10
	encrypted := make([]libunlynx.CipherText, len(tab))
	for i, v := range tab {
		encrypted[i] = *libunlynx.EncryptInt(pubKey, v)
	}

	protocol.TargetOfTransformation = encrypted
	protocol.Proofs = true
	feedback := protocol.FeedbackChannel
	protocol.Add = false
	protocol.KeyToRm = secKeyAddRm

	go protocol.Start()

	timeout := network.WaitRetry * time.Duration(network.MaxRetryConnect*5*2) * time.Millisecond

	select {
	case results := <-feedback:
		log.Lvl1("Results: ")
		log.Lvl1(results)
		decryptedResult := make([]int64, 2)
		for i, v := range results {
			decryptedResult[i] = libunlynx.DecryptInt(secKeyAfter, v)
		}
		assert.Equal(t, decryptedResult, expectedResults)
	case <-time.After(timeout):
		t.Fatal("Didn't finish in time")

	}
}
