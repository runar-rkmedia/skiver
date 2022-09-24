package backup

import (
	"bytes"
	"testing"
)

func TestCompress(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"Compress and decress text",
			args{
				b: []byte("Foo"),
			},
			3,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := &bytes.Buffer{}
			r := bytes.NewReader(tt.args.b)
			got, err := Compress(r, w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Compress() = '%v', want '%v'", got, tt.want)
			}
			rw := &bytes.Buffer{}
			rw.Write(w.Bytes())
			ww := &bytes.Buffer{}
			n, err := Decompress(rw, ww)

			if n != len(tt.args.b) {
				t.Errorf("Decompress() n = '%v', want '%v'", n, len(tt.args.b))
			}

			if ww.String() != string(tt.args.b) {
				t.Errorf("Decompress() = '%v', want '%v'", ww.String(), string(tt.args.b))
			}
		})
	}
}
