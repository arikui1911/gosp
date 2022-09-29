package gosp_test

import (
	"gosp"
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
        {"Quote", `'x`, gosp.List(gosp.Intern("QUOTE"), gosp.Intern("X"))},
        {"Quasiquote", "`x", gosp.List(gosp.Intern("QUASIQUOTE"), gosp.Intern("X"))},
        {"Unquote", `,x`, gosp.List(gosp.Intern("UNQUOTE"), gosp.Intern("X"))},
        {"Unquotesplicing", `,@x`, gosp.List(gosp.Intern("UNQUOTESPLICING"), gosp.Intern("X"))},
        {"List", `(+ 1 2)`, gosp.List(gosp.Intern("+"), gosp.IntegerValue(1), gosp.IntegerValue(2))},
        {"EmptyList", `()`, gosp.NilValue()},
        {"DotList", `(1 . 2)`, gosp.Cons(gosp.IntegerValue(1), gosp.IntegerValue(2))},
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
            /*
            if v != tt.want {
                t.Errorf("want %v(%T) got %v(%T)", tt.want, tt.want, v, v)
            }
            */
        })
    }

    /*
    r := gosp.NewReader(strings.NewReader(`123`))
    v, err := r.Read()
    if err != nil {
        t.Error(err)
        return
    }
    if v != gosp.IntegerValue(123) {
        t.Errorf("want %v(%T) got %v(%T)", 123, 123, v, v)
        return
    }
    */
}

