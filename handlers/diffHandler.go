package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"net/http"

	"github.com/jakobvarmose/go-qidenticon"
	"github.com/r3labs/diff/v2"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
)

func GetDiff(exportCache Cache) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		var input models.DiffSnapshotInput
		err := rc.ValidateBody(&input, false)
		if err != nil {
			return nil, err
		}

		if areEqaul(*input.A, *input.B) {
			return nil, NewApiError("Cannot diff with equal objects", http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
		}
		a, _, err := getExport(rc.L, exportCache, rc.Context.DB, importexport.ExportOptions{
			Project: *input.A.ProjectID,
			Tag:     input.A.Tag,
			Format:  input.Format,
		})
		if err != nil {
			return nil, err
		}
		b, _, err := getExport(rc.L, exportCache, rc.Context.DB, importexport.ExportOptions{
			Project: *input.B.ProjectID,
			Tag:     input.B.Tag,
			Format:  input.Format,
		})
		if err != nil {
			return nil, err
		}
		var changelog diff.Changelog
		aJson, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		bJson, err := json.Marshal(b)
		if err != nil {
			return nil, err
		}
		sizeA := len(aJson)
		sizeB := len(bJson)
		hA := hash256(aJson)
		hB := hash256(bJson)
		hashA := hex.EncodeToString(hA)
		hashB := hex.EncodeToString(hB)
		qiA := NewIdentiHash(hA)
		var qiB IdentiHash

		if hashA != hashB {

			c, err := diff.Diff(a, b, diff.DisableStructValues(), diff.AllowTypeMismatch(true))
			if err != nil {
				return nil, err
			}
			changelog = c
			qiB = NewIdentiHash(hB)
		}
		// return map[string]interface{}{"diff": changelog, "sizeA": sizeA, "sizeB": sizeB, "hashA": hashA, "hashB": hashB}, nil
		return DiffResponse{
			Diff: changelog,
			A: itemStats{
				Hash:       hashA,
				Size:       uint64(sizeA),
				IdentiHash: qiA.Img,
				ProjectID:  *input.A.ProjectID,
				Tag:        input.A.Tag,
			},
			B: itemStats{
				Hash:       hashB,
				Size:       uint64(sizeB),
				IdentiHash: qiB.Img,
				ProjectID:  *input.B.ProjectID,
				Tag:        input.B.Tag,
			},
		}, nil
	}
}

// swagger:response DiffResponse
type diffResponse struct {
	// In: body
	Data DiffResponse
}

type DiffResponse struct {
	Diff diff.Changelog `json:"diff"`
	A    itemStats      `json:"a"`
	B    itemStats      `json:"b"`
}

type itemStats struct {
	Hash       string `json:"hash"`
	Size       uint64 `json:"size"`
	IdentiHash []byte `json:"identi_hash"`
	ProjectID  string `json:"project_id"`
	Tag        string `json:"tag"`
}

func hash256(b []byte) []byte {
	hasher := sha256.New()
	hasher.Write(b)
	h := hasher.Sum(nil)
	return h
}

func getKey(s models.SnapshotSelector) string {
	return *s.ProjectID + s.Tag
}
func areEqaul(a, b models.SnapshotSelector) bool {
	return getKey(a) == getKey(b)
}

func NewIdentiHash(hash []byte) IdentiHash {
	code := qidenticon.Code(string(hash))
	img := qidenticon.Render(code, 70, qidenticon.DefaultSettings())
	var b bytes.Buffer
	// w := bufio.NewWriter(&b)

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
