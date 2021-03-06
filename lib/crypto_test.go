package libunlynx_test

import (
	"github.com/dedis/kyber"
	"github.com/dedis/kyber/util/random"
	"github.com/lca1/unlynx/lib"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

// TestNullCipherText verifies encryption, decryption and behavior of null ciphertexts.
func TestNullCipherText(t *testing.T) {

	secKey, pubKey := libunlynx.GenKey()

	nullEnc := libunlynx.EncryptInt(pubKey, 0)
	nullDec := libunlynx.DecryptInt(secKey, *nullEnc)

	if 0 != nullDec {
		t.Fatal("Decryption of encryption of 0 should be 0, got", nullDec)
	}

	var twoTimesNullEnc = libunlynx.CipherText{K: libunlynx.SuiTe.Point().Null(), C: libunlynx.SuiTe.Point().Null()}
	twoTimesNullEnc.Add(*nullEnc, *nullEnc)
	twoTimesNullDec := libunlynx.DecryptInt(secKey, twoTimesNullEnc)

	if 0 != nullDec {
		t.Fatal("Decryption of encryption of 0+0 should be 0, got", twoTimesNullDec)
	}

}

// TestEncryption tests a relatively high number of encryptions.
func TestEncryption(t *testing.T) {

	_, pubKey := libunlynx.GenKey()

	nbrEncryptions := 2
	for i := 0; i < nbrEncryptions; i++ {
		libunlynx.EncryptInt(pubKey, 0)
	}
}

// TestDecryptionConcurrent test the multiple encryptions/decryptions at the same time
func TestDecryptionConcurrent(t *testing.T) {
	numThreads := 5

	sec, pubKey := libunlynx.GenKey()

	libunlynx.StartParallelize(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			ct := libunlynx.EncryptInt(pubKey, 0)
			val := libunlynx.DecryptInt(sec, *ct)
			assert.Equal(t, val, int64(0))
		}()
	}
}

// TestDecryptionConcurrent test the multiple encryptions/decryptions at the same time
func TestDecryptionNegConcurrent(t *testing.T) {
	numThreads := 5

	sec, pubKey := libunlynx.GenKey()

	libunlynx.StartParallelize(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			ct := libunlynx.EncryptInt(pubKey, 3)
			val := libunlynx.DecryptIntWithNeg(sec, *ct)
			assert.Equal(t, val, int64(3))

			ct = libunlynx.EncryptInt(pubKey, 3)
			val = libunlynx.DecryptInt(sec, *ct)
			assert.Equal(t, val, int64(3))

			ct = libunlynx.EncryptInt(pubKey, -3)
			val = libunlynx.DecryptIntWithNeg(sec, *ct)
			assert.Equal(t, val, int64(-3))
		}()
	}
}

// TestNullCipherText verifies encryption, decryption and behavior of null cipherVectors.
func TestNullCipherVector(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()

	nullVectEnc := *libunlynx.NullCipherVector(10, pubKey)
	nullVectDec := libunlynx.DecryptIntVector(secKey, &nullVectEnc)

	target := []int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if !reflect.DeepEqual(nullVectDec, target) {
		t.Fatal("Null vector of dimension 4 should be ", target, "got", nullVectDec)
	}

	twoTimesNullEnc := libunlynx.NewCipherVector(10)
	twoTimesNullEnc.Add(nullVectEnc, nullVectEnc)
	twoTimesNullDec := libunlynx.DecryptIntVector(secKey, twoTimesNullEnc)

	if !reflect.DeepEqual(twoTimesNullDec, target) {
		t.Fatal("Null vector + Null vector should be ", target, "got", twoTimesNullDec)
	}
}

// TestHomomorphicOpp tests homomorphic addition.
func TestHomomorphicOpp(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()

	cv1 := libunlynx.EncryptIntVector(pubKey, []int64{0, 1, 2, 3, 100})
	cv2 := libunlynx.EncryptIntVector(pubKey, []int64{0, 0, 1, 3, 3})
	targetAdd := []int64{0, 1, 3, 6, 103}
	targetSub := []int64{0, 1, 1, 0, 97}
	targetMul := int64(4)

	cv3 := libunlynx.NewCipherVector(5)
	cv3.Add(*cv1, *cv2)
	cv4 := libunlynx.NewCipherVector(5)
	cv4.Sub(*cv1, *cv2)
	cv5 := libunlynx.EncryptInt(pubKey, 2)
	cv5.MulCipherTextbyScalar(*cv5, libunlynx.SuiTe.Scalar().SetInt64(2))

	pAdd := libunlynx.DecryptIntVector(secKey, cv3)
	pSub := libunlynx.DecryptIntVector(secKey, cv4)
	pMul := libunlynx.DecryptInt(secKey, *cv5)

	assert.Equal(t, targetAdd, pAdd)
	assert.Equal(t, targetSub, pSub)
	assert.Equal(t, targetMul, pMul)
}

// TestCryptoTagging tests the deterministic tagging
func TestCryptoTagging(t *testing.T) {
	const N = 5

	groupKey, private, _ := libunlynx.GenKeys(N)
	_, secretPrivate, _ := libunlynx.GenKeys(N)

	target := []int64{-8358645081376817152, -8358645081376817152, 2, 3, 2, 5}
	cv := *libunlynx.EncryptIntVector(groupKey, target)
	for n := 0; n < N; n++ {
		tmp := libunlynx.NewCipherVector(len(cv))
		tmp.DeterministicTagging(&cv, private[n], secretPrivate[n])

		cv = *tmp

	}
	assert.True(t, cv[0].C.Equal(cv[1].C))
	assert.True(t, cv[2].C.Equal(cv[4].C))
	assert.False(t, cv[0].C.Equal(cv[3].C))
}

// TestCryptoKeySwitching tests key switching.
func TestCryptoKeySwitching(t *testing.T) {
	const N = 5
	groupKey, privates, _ := libunlynx.GenKeys(N)
	newPrivate, newPublic := libunlynx.GenKey()

	target := []int64{1, 2, 3, 4, 5}
	cv := libunlynx.EncryptIntVector(groupKey, target)

	origEphem := make([]kyber.Point, len(*cv))
	kscv := make(libunlynx.CipherVector, len(*cv))
	for i, c := range *cv {
		origEphem[i] = c.K
		kscv[i].K = libunlynx.SuiTe.Point().Null()
		kscv[i].C = c.C
	}

	for n := 0; n < N; n++ {
		kscv.KeySwitching(kscv, origEphem, newPublic, privates[n])
	}

	res := libunlynx.DecryptIntVector(newPrivate, &kscv)
	assert.True(t, reflect.DeepEqual(res, target))

}

// TestEqualDeterministCipherText tests equality between deterministic ciphertexts.
func TestEqualDeterministCipherText(t *testing.T) {
	dcv1 := libunlynx.DeterministCipherVector{libunlynx.DeterministCipherText{Point: libunlynx.SuiTe.Point().Base()}, libunlynx.DeterministCipherText{Point: libunlynx.SuiTe.Point().Null()}}
	dcv2 := libunlynx.DeterministCipherVector{libunlynx.DeterministCipherText{Point: libunlynx.SuiTe.Point().Base()}, libunlynx.DeterministCipherText{Point: libunlynx.SuiTe.Point().Null()}}

	assert.True(t, dcv1.Equal(&dcv2))
	assert.True(t, dcv1.Equal(&dcv1))

	dcv1 = libunlynx.DeterministCipherVector{}
	dcv2 = libunlynx.DeterministCipherVector{}
	assert.True(t, dcv1.Equal(&dcv2))
	assert.True(t, dcv1.Equal(&dcv1))

	var nilp *libunlynx.DeterministCipherVector
	pdcv1 := &dcv1
	assert.True(t, pdcv1.Equal(&dcv2))
	assert.False(t, pdcv1.Equal(nilp))

	pdcv1 = nil
	assert.False(t, pdcv1.Equal(&dcv2))
	assert.True(t, pdcv1.Equal(nilp))
}

// TestAbstractPointsConverter tests the kyber points array converter (to bytes)
func TestAbstractPointsConverter(t *testing.T) {
	aps := make([]kyber.Point, 0)

	clientPrivate := libunlynx.SuiTe.Scalar().Pick(random.New())

	for i := 0; i < 4; i++ {
		ap := libunlynx.SuiTe.Point().Mul(clientPrivate, libunlynx.SuiTe.Point().Base())
		aps = append(aps, ap)
	}

	apsBytes := libunlynx.AbstractPointsToBytes(aps)
	newAps := libunlynx.BytesToAbstractPoints(apsBytes)

	for i, el := range aps {
		if !reflect.DeepEqual(el.String(), newAps[i].String()) {
			t.Fatal("Wrong results, expected", el, "but got", newAps[i])
		}
	}

	t.Log("[AbstractPoints] -> Good results")
}

// TestCiphertextConverter tests the Ciphertext converter (to bytes)
func TestCiphertextConverter(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()

	target := int64(2)
	ct := libunlynx.EncryptInt(pubKey, target)

	ctb := ct.ToBytes()

	newCT := libunlynx.CipherText{}
	newCT.FromBytes(ctb)

	p := libunlynx.DecryptInt(secKey, newCT)

	assert.Equal(t, target, p)
}

// TestCipherVectorConverter tests the CipherVector converter (to bytes)
func TestCipherVectorConverter(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()

	target := []int64{0, 1, 3, 103, 103}
	cv := libunlynx.EncryptIntVector(pubKey, target)

	cvb, length := cv.ToBytes()

	newCV := libunlynx.CipherVector{}
	newCV.FromBytes(cvb, length)

	p := libunlynx.DecryptIntVector(secKey, &newCV)

	assert.Equal(t, target, p)
}

// TestIntArrayToCipherVector tests the int array to CipherVector converter and IntToPoint + PointToCiphertext
func TestIntArrayToCipherVector(t *testing.T) {
	integers := []int64{1, 2, 3, 4, 5, 6}

	cipherVect := libunlynx.IntArrayToCipherVector(integers)
	for i, v := range cipherVect {
		B := libunlynx.SuiTe.Point().Base()
		i := libunlynx.SuiTe.Scalar().SetInt64(integers[i])
		M := libunlynx.SuiTe.Point().Mul(i, B)
		N := libunlynx.SuiTe.Point().Null()
		assert.Equal(t, v.C, M)
		assert.Equal(t, v.K, N)
	}
}

func TestB64Serialization(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()
	target := []int64{0, 1, 3, 103, 103}
	cv := libunlynx.EncryptIntVector(pubKey, target)

	for i, ct := range *cv {
		ctSerialized := ct.Serialize()

		// with newciphertext
		ctDeserialized := libunlynx.NewCipherTextFromBase64(ctSerialized)
		decVal := libunlynx.DecryptInt(secKey, *ctDeserialized)
		assert.Equal(t, target[i], decVal)

		// with deserialize
		ctDeserializedBis := libunlynx.NewCipherText()
		ctDeserializedBis.Deserialize(ctSerialized)
		decValBis := libunlynx.DecryptInt(secKey, *ctDeserializedBis)
		assert.Equal(t, target[i], decValBis)
		assert.Equal(t, decVal, decValBis)
	}
}

func TestEncryptScalar(t *testing.T) {
	target := int64(5)
	secKey, pubKey := libunlynx.GenKey()
	scal := libunlynx.SuiTe.Scalar().SetInt64(target)

	ct := *libunlynx.EncryptScalar(pubKey, scal)
	assert.Equal(t, libunlynx.DecryptInt(secKey, ct), target)
}

func TestEncryptScalarVector(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()
	target := []int64{0, 1, 3, 103, 103}

	targetScal := make([]kyber.Scalar, len(target))
	for i, v := range target {
		targetScal[i] = libunlynx.SuiTe.Scalar().SetInt64(v)
	}

	cv := *libunlynx.EncryptScalarVector(pubKey, targetScal)

	for i, v := range cv {
		assert.Equal(t, libunlynx.DecryptInt(secKey, v), target[i])
	}
}

func TestDecryptIntVectorWithNeg(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()
	target := []int64{0, -1, -3, -103, -103}
	cv := libunlynx.EncryptIntVector(pubKey, target)

	results := libunlynx.DecryptIntVectorWithNeg(secKey, cv)
	for i, v := range results {
		assert.Equal(t, target[i], v)
	}
}

func TestDecryptCheckZero(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()
	ctZero := *libunlynx.EncryptInt(pubKey, int64(0))
	assert.Equal(t, libunlynx.DecryptCheckZero(secKey, ctZero), int64(0))

	ctNotZero := *libunlynx.EncryptInt(pubKey, int64(5))
	assert.Equal(t, libunlynx.DecryptCheckZero(secKey, ctNotZero), int64(1))
}

func TestDecryptCheckZeroVector(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()
	target := []int64{0, 1, 5, 0, 3}

	cv := libunlynx.EncryptIntVector(pubKey, target)
	for i, v := range libunlynx.DecryptCheckZeroVector(secKey, cv) {
		if i == 0 || i == 3 {
			assert.Equal(t, v, int64(0))
		} else {
			assert.Equal(t, v, int64(1))
		}
	}
}

func TestNewDeterministicCipherText(t *testing.T) {
	secKey, _ := libunlynx.GenKey()
	dt1 := libunlynx.NewDeterministicCipherText()
	dt2 := libunlynx.NewDeterministicCipherText()
	ctTest := libunlynx.CipherText{
		K: dt1.Point,
		C: dt2.Point,
	}

	assert.Equal(t, libunlynx.DecryptInt(secKey, ctTest), int64(0))
}

func TestNewDeterministicCipherVector(t *testing.T) {
	secKey, _ := libunlynx.GenKey()
	k := 5
	dtv1 := *libunlynx.NewDeterministicCipherVector(k)
	dtv2 := *libunlynx.NewDeterministicCipherVector(k)
	cv := make(libunlynx.CipherVector, k)

	for i := range cv {
		cv[i] = libunlynx.CipherText{
			K: dtv1[i].Point,
			C: dtv2[i].Point,
		}
	}

	test := libunlynx.DecryptIntVector(secKey, &cv)
	for _, v := range test {
		assert.Equal(t, v, int64(0))
	}
}

func TestDeterministicCipherTextKey(t *testing.T) {
	_, pubKey := libunlynx.GenKey()
	k := 5
	target := []int64{0, 1, 3, 103, 103}
	dcv := make(libunlynx.DeterministCipherVector, k)
	str := make([]string, k)
	for i := range dcv {
		dcv[i] = libunlynx.DeterministCipherText{Point: (*libunlynx.EncryptInt(pubKey, target[i])).C}
		str[i] = dcv[i].Point.String()
	}

	assert.Equal(t, libunlynx.GroupingKey(strings.Join(str, "")), dcv.Key())
}

func TestSerializePoint(t *testing.T) {
	secKey, pubKey := libunlynx.GenKey()
	target := int64(1)
	ct := *libunlynx.EncryptInt(pubKey, target)
	strC, errC1 := libunlynx.SerializePoint(ct.C)
	strK, errK1 := libunlynx.SerializePoint(ct.K)
	if errC1 != nil || errK1 != nil {
		t.Fatal("Error in the serialization")
	}

	pointC, errC2 := libunlynx.DeserializePoint(strC)
	pointK, errK2 := libunlynx.DeserializePoint(strK)
	if errC2 != nil || errK2 != nil {
		t.Fatal("Error in the serialization")
	}

	ct.C = pointC
	ct.K = pointK

	assert.Equal(t, target, libunlynx.DecryptInt(secKey, ct))
}

func TestSerializeScalar(t *testing.T) {
	target := int64(5)
	sclr := libunlynx.SuiTe.Scalar().SetInt64(target)
	str, err1 := libunlynx.SerializeScalar(sclr)
	sclrTest, err2 := libunlynx.DeserializeScalar(str)

	if err1 != nil || err2 != nil {
		t.Fatal("Error in the serialization")
	}

	assert.Equal(t, sclr, sclrTest)
}
