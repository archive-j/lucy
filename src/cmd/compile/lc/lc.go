package lc

import (
	"fmt"
	"io/ioutil"

	"github.com/756445638/lucy/src/cmd/compile/parser"
)

func Main(files []string) {
	l := &LucyCompile{
		Files:            files,
		NerrsStopCompile: 10,
		Nerrs:            []error{},
	}
	l.compile()
}

type LucyCompile struct {
	Files            []string
	Nerrs            []error
	NerrsStopCompile int
}

func (l *LucyCompile) exit() {
	for _, v := range l.Nerrs {
		fmt.Println(v)
	}
}

func (l *LucyCompile) compile() {
	for _, v := range l.Files {
		bs, err := ioutil.ReadFile(v)
		if err != nil {
			l.Nerrs = append(l.Nerrs, err)
			continue
		}
		l.Nerrs = append(l.Nerrs, parser.Parse(&Tops, v, bs, CompileFlags.OnlyImport)...)

		if len(l.Nerrs) > 10 {
			l.exit()
		}
	}

}
