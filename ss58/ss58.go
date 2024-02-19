package ss58

import (
	"strings"

	"github.com/itering/subscan/util"
	"github.com/itering/subscan/util/base58"
	"golang.org/x/crypto/blake2b"
)

const SubspaceAddressType = 2254

func Decode(address string, addressType int) string {
	checksumPrefix := []byte("SS58PRE")
	ss58Format := base58.Decode(address)
	// if len(ss58Format) == 0 || ss58Format[0] != byte(addressType) {
	// 	return ""
	// }
	if len(ss58Format) == 0 {
		return ""
	}
	var checksumLength int
	if util.IntInSlice(len(ss58Format), []int{3, 4, 6, 10}) {
		checksumLength = 1
	} else if util.IntInSlice(len(ss58Format), []int{5, 7, 11, 35, 36}) {
		checksumLength = 2
	} else if util.IntInSlice(len(ss58Format), []int{8, 12}) {
		checksumLength = 3
	} else if util.IntInSlice(len(ss58Format), []int{9, 13}) {
		checksumLength = 4
	} else if util.IntInSlice(len(ss58Format), []int{14}) {
		checksumLength = 5
	} else if util.IntInSlice(len(ss58Format), []int{15}) {
		checksumLength = 6
	} else if util.IntInSlice(len(ss58Format), []int{16}) {
		checksumLength = 7
	} else if util.IntInSlice(len(ss58Format), []int{17}) {
		checksumLength = 8
	} else {
		return ""
	}
	bss := ss58Format[0 : len(ss58Format)-checksumLength]
	checksum, _ := blake2b.New(64, []byte{})
	w := append(checksumPrefix[:], bss[:]...)
	_, err := checksum.Write(w)
	if err != nil {
		return ""
	}

	h := checksum.Sum(nil)
	if util.BytesToHex(h[0:checksumLength]) != util.BytesToHex(ss58Format[len(ss58Format)-checksumLength:]) {
		return ""
	}
	return util.BytesToHex(ss58Format[len(ss58Format)-34 : len(ss58Format)-checksumLength])
}

func Encode(address string, prefix int) string {
	checksumPrefix := []byte("SS58PRE")
	addressBytes := []byte(address)
	if strings.HasPrefix(address, "0x") {
		addressBytes = util.HexToBytes(address)
	}

	var checksumLength int
	if len(addressBytes) == 32 {
		checksumLength = 2
	} else if util.IntInSlice(len(addressBytes), []int{1, 2, 4, 8}) {
		checksumLength = 1
	} else {
		return ""
	}

	simplePrefix := prefix & 0x3F
	fullPrefix := 0x4000 | ((prefix >> 8) & 0x3F) | ((prefix & 0xFF) << 6)
	prefixHigh := fullPrefix >> 8
	prefixLow := fullPrefix & 0xFF

	prefixBytes := make([]byte, 0)
	if prefix == simplePrefix {
		prefixBytes = append(prefixBytes, byte(simplePrefix))
	} else {
		prefixBytes = append(prefixBytes, byte(prefixHigh), byte(prefixLow))
	}

	addressFormat := append(prefixBytes[:], addressBytes[:]...)
	checksum, _ := blake2b.New(64, []byte{})
	w := append(checksumPrefix[:], addressFormat[:]...)
	_, err := checksum.Write(w)
	if err != nil {
		return ""
	}

	h := checksum.Sum(nil)
	b := append(addressFormat[:], h[:checksumLength][:]...)
	return base58.Encode(b)
}
