package sourcemap

import (
	"strconv"
	"strings"
)

type JsonPath struct {
	path  []string
	Queue string
}

func (jp *JsonPath) SetQueue(s string) {
	jp.Queue = s
}
func (jp *JsonPath) IsWithinArray() (int, bool) {
	if len(jp.path) == 0 {
		return -1, false
	}
	last := jp.path[len(jp.path)-1]
	i, err := strconv.ParseInt(last, 10, 64)
	if err != nil {
		return -1, false
	}
	return int(i), true
}
func (jp *JsonPath) AddArrayElement() {
	i, ok := jp.IsWithinArray()
	if !ok {
		jp.path = append(jp.path, "0")
		return
	}
	jp.path[len(jp.path)-1] = strconv.FormatInt(int64(i+1), 10)
}
func (jp *JsonPath) Add(s string) {
	jp.path = append(jp.path, s)
}
func (jp *JsonPath) String() string {
	var s []string
	s = append(s, jp.path...)
	if jp.Queue != "" {
		s = append(jp.path, jp.Queue)
	}
	return strings.Join(s, ".")
}

func NewJsonPath() JsonPath {
	return JsonPath{path: []string{}}
}
