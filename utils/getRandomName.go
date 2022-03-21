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
func GetRandomName() (string, error) {
	var result []string
	bs := make([]byte, 4)
	_, err := rand.Read(bs)
	if err != nil {
		return "", err
	}
	result = mnemonicode.EncodeWordList(result, bs)
	return GenerateRandomPin() + "-" + strings.Join(result, "-"), nil
}

func HashName(b []byte) string {
	var result []string
	words := mnemonicode.EncodeWordList(result, b)
	return strings.Join(words, "-")
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
