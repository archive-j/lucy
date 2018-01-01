package jvm

import (
	"encoding/binary"

	"github.com/756445638/lucy/src/cmd/compile/jvm/cg"
)

func mkPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "$" + name
}

func appendBackPatch(p *[][]byte, b []byte) {
	if *p == nil {
		*p = [][]byte{b}
	} else {
		*p = append(*p, b)
	}
}
func backPatchEs(es [][]byte, code *cg.AttributeCode) {
	for _, v := range es {
		binary.BigEndian.PutUint16(v, code.CodeLength)
	}
}
