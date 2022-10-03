package gosp_test

import (
	"fmt"
	"gosp"
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReader(t *testing.T) {
    tests := []struct {
        name string
        src string
        want gosp.Value
    }{
        {"EOF", ``, nil},
        {"Int", `123`, gosp.IntegerValue(123)},
        {"Float", `4.56`, gosp.FloatValue(4.56)},
        {"Symbol", `hoge`, gosp.Intern("HOGE")},
        {"String", `"Hello, world!"`, gosp.NewString("Hello, world!")},
        {"EscapeSeq", `"\a\b\t\n\v\f\r\"\\"`, gosp.NewString("\a\b\t\n\v\f\r\"\\")},
        {"InvalidEscSeq", `"\c"`, gosp.NewString("c")},
        {"Comment", `123  ; 456`, gosp.IntegerValue(123)},
        {"Quote", `'x`, gosp.List(gosp.Intern("QUOTE"), gosp.Intern("X"))},
        {"Quasiquote", "`x", gosp.List(gosp.Intern("QUASIQUOTE"), gosp.Intern("X"))},
        {"Unquote", `,x`, gosp.List(gosp.Intern("UNQUOTE"), gosp.Intern("X"))},
        {"Unquotesplicing", `,@x`, gosp.List(gosp.Intern("UNQUOTESPLICING"), gosp.Intern("X"))},
        {"List", `(+ 1 2)`, gosp.List(gosp.Intern("+"), gosp.IntegerValue(1), gosp.IntegerValue(2))},
        {"EmptyList", `()`, gosp.NilValue()},
        {"DotList", `(1 . 2)`, gosp.Cons(gosp.IntegerValue(1), gosp.IntegerValue(2))},
        {"AtomDelimNL", "x\n", gosp.Intern("X")},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T){
            r := gosp.NewReader(strings.NewReader(tt.src))
            v, err := r.Read()
            if err != nil {
                t.Error(err)
                return
            }
            if diff := cmp.Diff(tt.want, v); diff != "" {
                t.Errorf("want <-> got <+>\n%s", diff)
            }
        })
    }
}

func TestReaderError(t *testing.T) {
    tests := []struct{
        name string
        src string
    }{
        {"syntaxerror: ')'", `)`},
        {"syntaxerror: '.'", `.`},
        {"toobigint", fmt.Sprintf("%d1", math.MaxInt64)},
        {"toobigfloat", fmt.Sprintf("1%f", math.MaxFloat64)},
        {"untermstring", `"Hello`},
        {"untermstringesc", `"Hoge\`},
        {"errwithquote", `')`},
        {"invaliddotpair", `(1 . 2 3)`},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T){
            r := gosp.NewReader(strings.NewReader(tt.src))
            _, err := r.Read()
            if err == nil {
                t.Error("want error returned but nil")
            }
        })
    }
}

