package libunlynxtools

import (
	"encoding/gob"
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb/errors"
	"github.com/dedis/kyber"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/lca1/unlynx/lib"
	"os"
	"strconv"
	"strings"
)

// SendISMOthers sends a message to all other services
func SendISMOthers(s *onet.ServiceProcessor, el *onet.Roster, msg interface{}) error {
	var errStrs []string
	for _, e := range el.List {
		if !e.ID.Equal(s.ServerIdentity().ID) {
			log.Lvl3("Sending to", e)
			err := s.SendRaw(e, msg)
			if err != nil {
				errStrs = append(errStrs, err.Error())
			}
		}
	}
	var err error
	if len(errStrs) > 0 {
		err = errors.New(strings.Join(errStrs, "\n"))
	}
	return err
}

// AddInMap permits to add a filtered response with its deterministic tag in a map
func AddInMap(s map[libunlynx.GroupingKey]libunlynx.FilteredResponse, key libunlynx.GroupingKey, added libunlynx.FilteredResponse) {
	if localResult, ok := s[key]; !ok {
		s[key] = added
	} else {
		tmp := libunlynx.NewFilteredResponse(len(added.GroupByEnc), len(added.AggregatingAttributes))
		s[key] = *tmp.Add(localResult, added)
	}
}

// Int64ArrayToString transforms an integer array into a string
func Int64ArrayToString(s []int64) string {
	if len(s) == 0 {
		return ""
	}

	result := ""
	for _, elem := range s {
		result += fmt.Sprintf("%v ", elem)
	}
	return result
}

// StringToInt64Array transforms a string ("1 0 1 0") to an integer array
func StringToInt64Array(s string) []int64 {
	if len(s) == 0 {
		return make([]int64, 0)
	}

	container := strings.Split(s, " ")

	result := make([]int64, 0)
	for _, elem := range container {
		if elem != "" {
			aux, _ := strconv.ParseInt(elem, 10, 64)
			result = append(result, aux)
		}
	}
	return result
}

// ConvertDataToMap a converts an array of integers to a map of id -> integer
func ConvertDataToMap(data []int64, first string, start int) map[string]int64 {
	result := make(map[string]int64)
	for _, el := range data {
		result[first+strconv.Itoa(start)] = el
		start++
	}
	return result
}

// ConvertMapToData converts the map into a slice of int64 (to ease out printing and aggregation)
func ConvertMapToData(data map[string]int64, first string, start int) []int64 {
	result := make([]int64, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[first+strconv.Itoa(start)]
		start++
	}
	return result
}

// WriteToGobFile stores object (e.g. lib.Enc_CipherVectorScalar) in a gob file. Note that the object must contain serializable stuff, for example byte arrays.
func WriteToGobFile(path string, object interface{}) {
	file, err := os.Create(path)
	defer file.Close()

	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	} else {
		log.Fatal("Could not write Gob file: ", err)
	}
}

// ReadFromGobFile reads data from gob file to the object
func ReadFromGobFile(path string, object interface{}) {
	file, err := os.Open(path)
	defer file.Close()

	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	} else {
		log.Fatal("Could not read Gob file: ", err)
	}
}

// EncodeCipherVectorScalar converts the data inside lib.CipherVectorScalar to bytes and stores it in a new object to be saved in the gob file
func EncodeCipherVectorScalar(cV []libunlynx.CipherVectorScalar) ([]libunlynx.CipherVectorScalarBytes, error) {
	slice := make([]libunlynx.CipherVectorScalarBytes, 0)

	for _, v := range cV {
		eCV := libunlynx.CipherVectorScalarBytes{}

		for _, el := range v.S {
			scalar, err := el.MarshalBinary()

			if err != nil {
				return slice, err
			}

			eCV.S = append(eCV.S, scalar)
		}

		for _, el := range v.CipherV {
			container := make([][]byte, 0)

			c, err := el.C.MarshalBinary()

			if err != nil {
				return slice, err
			}

			k, err := el.K.MarshalBinary()

			if err != nil {
				return slice, err
			}

			container = append(container, k, c)

			eCV.CipherV = append(eCV.CipherV, container)
		}

		slice = append(slice, eCV)
	}

	return slice, nil
}

// DecodeCipherVectorScalar converts the byte data stored in the lib.Enc_CipherVectorScalar (which is read from the gob file) to a new lib.CipherVectorScalar
func DecodeCipherVectorScalar(eCV []libunlynx.CipherVectorScalarBytes) ([]libunlynx.CipherVectorScalar, error) {
	slice := make([]libunlynx.CipherVectorScalar, 0)

	for _, v := range eCV {
		cV := libunlynx.CipherVectorScalar{}

		for _, el := range v.S {
			s := libunlynx.SuiTe.Scalar()
			if err := s.UnmarshalBinary(el); err != nil {
				return slice, err
			}

			cV.S = append(cV.S, s)
		}

		for _, el := range v.CipherV {
			k := libunlynx.SuiTe.Point()
			if err := k.UnmarshalBinary(el[0]); err != nil {
				return slice, err
			}

			c := libunlynx.SuiTe.Point()
			if err := c.UnmarshalBinary(el[1]); err != nil {
				return slice, err
			}

			cipher := libunlynx.CipherText{K: k, C: c}
			cV.CipherV = append(cV.CipherV, cipher)

		}

		slice = append(slice, cV)
	}

	return slice, nil
}

// JoinAttributes joins clear and encrypted attributes into one encrypted container (CipherVector)
func JoinAttributes(clear, enc map[string]int64, identifier string, encryptionKey kyber.Point) libunlynx.CipherVector {
	clearContainer := ConvertMapToData(clear, identifier, 0)
	encContainer := ConvertMapToData(enc, identifier, len(clear))

	result := make(libunlynx.CipherVector, 0)

	for i := 0; i < len(clearContainer); i++ {
		result = append(result, *libunlynx.EncryptInt(encryptionKey, int64(clearContainer[i])))
	}
	for i := 0; i < len(encContainer); i++ {
		result = append(result, *libunlynx.EncryptInt(encryptionKey, int64(encContainer[i])))
	}

	return result
}

// FromDpClearResponseToProcess converts a DpClearResponse struct to a ProcessResponse struct
func FromDpClearResponseToProcess(dcr *libunlynx.DpClearResponse, encryptionKey kyber.Point) libunlynx.ProcessResponse {
	result := libunlynx.ProcessResponse{}

	result.AggregatingAttributes = JoinAttributes(dcr.AggregatingAttributesClear, dcr.AggregatingAttributesEnc, "s", encryptionKey)
	result.WhereEnc = JoinAttributes(dcr.WhereClear, dcr.WhereEnc, "w", encryptionKey)
	result.GroupByEnc = JoinAttributes(dcr.GroupByClear, dcr.GroupByEnc, "g", encryptionKey)

	return result
}
