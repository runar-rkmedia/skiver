package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"

	"github.com/dustin/go-humanize"
	"github.com/jakobvarmose/go-qidenticon"
)

func NewBinaryMeta(a interface{}) (*BinaryMeta, error) {
	aJson, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	bm := NewBinaryMetaFromBytes(aJson)
	return &bm, nil
}

// For consistency, input is expectd to be the return-type of json.Marshal(input), but works with any []byte
func NewBinaryMetaFromBytes(b []byte) BinaryMeta {
	hA := hash256(b)
	bm := BinaryMeta{
		Size:       uint64(len(b)),
		Hash:       hex.EncodeToString(hA),
		IdentiHash: NewIdentiHash(hA).Img,
	}
	bm.SizeHumanized = humanize.Bytes(bm.Size)
	return bm
}

// BinaryMeta is a humanized version of the raw data of an object
type BinaryMeta struct {
	Hash          string `json:"hash"`
	Size          uint64 `json:"size"`
	SizeHumanized string `json:"size_humanized"`
	IdentiHash    []byte `json:"identi_hash"`
}

func (bm BinaryMeta) Equal(bm2 BinaryMeta) bool {
	return bm.Hash == bm2.Hash
}

func hash256(b []byte) []byte {
	hasher := sha256.New()
	hasher.Write(b)
	h := hasher.Sum(nil)
	return h
}

func NewIdentiHash(hash []byte) IdentiHash {
	code := qidenticon.Code(string(hash))
	img := qidenticon.Render(code, 72, qidenticon.DefaultSettings())
	var b bytes.Buffer

	err := png.Encode(&b, img)
	if err != nil {
		panic(err)
	}
	return IdentiHash{
		Hash: hash,
		Img:  b.Bytes(),
	}

}

type IdentiHash struct {
	Img  []byte
	Hash []byte
}
