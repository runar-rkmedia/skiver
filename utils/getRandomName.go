package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/schollz/mnemonicode"
)

// Attribution: https://github.com/schollz/croc/blob/master/src/utils/utils.go

// GetRandomName returns mnemonicoded random name
func GetRandomName() string {
	var result []string
	bs := make([]byte, 4)
	rand.Read(bs)
	result = mnemonicode.EncodeWordList(result, bs)
	return GenerateRandomPin() + "-" + strings.Join(result, "-")
}

func GenerateRandomPin() string {
	s := ""
	max := new(big.Int)
	max.SetInt64(9)
	for i := 0; i < 4; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		s += fmt.Sprintf("%d", v)
	}
	return s
}
