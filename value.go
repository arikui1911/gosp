package gosp

type Value interface{}

type Symbol struct {
    Name string
    Value Value
    Function Value
}

func Intern(name string) *Symbol {
    return &Symbol {
        Name: name,
    }
}

type Nil struct{}

func NilValue() Nil {
    return Nil{}
}

type Cell struct {
    Car Value
    Cdr Value
}

func Cons(car Value, cdr Value) *Cell {
    return &Cell{
        Car: car,
        Cdr: cdr,
    }
}

func List(elements... Value) Value {
    if (len(elements) == 0) {
        return NilValue()
    }
    return Cons(elements[0], List(elements[1:]...))
}

type String string

func NewString(s string) String {
    return String(s)
}

type Integer int64

func IntegerValue(n int64) Integer {
    return Integer(n)
}

type Float float64

func FloatValue(x float64) Float {
    return Float(x)
}

