package gosp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

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

type Reader struct {
    src *bufio.Reader
    ch chan scanResult
    lineno int
    column int

    // tokenのバッファリング用
    savedToken token
    hasSavedToken bool

    // runeのバッファリング用
    savedRune rune
    hasSavedRune bool
    lastNewlineColumn int
}

func NewReader(src io.Reader) *Reader {
    r := &Reader{
        src: bufio.NewReader(src),
        ch: make(chan scanResult, 1),
        lineno: 1,
        column: 0,
    }
    go scan(r)
    return r
}

func (r *Reader) Read() (Value, error) {
    t, err := r.nextToken()
    if err != nil {
        return nil, err
    }
    switch t.tag {
    case token_EOF:
        return nil, nil
    case token_LP:
        return readList(r)
    case token_QUOTE:
        return readQuote(r, "QUOTE")
    case token_BACKQUOTE:
        return readQuote(r, "QUASIQUOTE")
    case token_COMMA:
        return readQuote(r, "UNQUOTE")
    case token_COMMAAT:
        return readQuote(r, "UNQUOTESPLICING")
    case token_STRING:
        return NewString(t.value), nil
    case token_ATOM:
        return parseAtom(t)
    }
    return nil, fmt.Errorf("%d:%d: unexpected token", t.lineno, t.column)
}

func readList(r *Reader) (Value, error) {
    t, err := r.nextToken()
    if err != nil {
        return nil, err
    }
    if t.tag == token_RP {
        return NilValue(), nil
    }
    r.pushbackToken(t)

    car, err := r.Read()
    if err != nil {
        return nil, err
    }

    t, err = r.nextToken()
    if err != nil {
        return nil, err
    }
    if t.tag == token_DOT {
        cdr, err := r.Read()
        if err != nil {
            return nil, err
        }
        t, err := r.nextToken()
        if err != nil {
            return nil, err
        }
        if t.tag != token_RP {
            return nil, fmt.Errorf("%d:%d: unexpeted token", t.lineno, t.column)
        }
        return Cons(car, cdr), nil
    }
    r.pushbackToken(t)

    cdr, err := readList(r)
    if err != nil {
        return nil, err
    }
    return Cons(car, cdr), nil
}

func readQuote(r *Reader, quoteName string) (Value, error) {
    exp, err := r.Read()
    if err != nil {
        return nil, err
    }
    return Cons(Intern(quoteName), Cons(exp, NilValue())), nil
}

func parseAtom(t token) (Value, error) {
    i64, err := strconv.ParseInt(t.value, 10, 64)
    if err == nil {
        return IntegerValue(i64), nil
    }
    if !isStrconvSyntaxError(err) {
        return nil, err
    }

    f64, err := strconv.ParseFloat(t.value, 64)
    if err == nil {
        return FloatValue(f64), nil
    }
    if !isStrconvSyntaxError(err) {
        return nil, err
    }

    return Intern(strings.ToUpper(t.value)), nil
}

func isStrconvSyntaxError(err error) bool {
    ne, ok := err.(*strconv.NumError)
    if !ok {
        return false
    }
    return ne.Err == strconv.ErrSyntax
}

func (r *Reader) nextToken() (token, error) {
    if r.hasSavedToken {
        r.hasSavedToken = false
        return r.savedToken, nil
    }
    sr := <-r.ch
    return sr.tok, sr.err
}

func (r *Reader) pushbackToken(t token) {
    r.savedToken = t
    r.hasSavedToken = true
}

func (r *Reader) readRune() (rune, bool) {
    var c rune
    var err error
    if r.hasSavedRune {
        r.hasSavedRune = false
        c = r.savedRune
    } else {
        c, _, err = r.src.ReadRune()
    }
    if c == '\n' {
        r.lastNewlineColumn = r.column
        r.lineno++
        r.column = 0
    } else {
        r.column++
    }
    if err == io.EOF {
        return c, false
    }
    if err != nil {
        r.ch <- scanResult{err: err}
        return c, false
    }
    return c, true
}

func (r *Reader) unreadRune(c rune) {
    r.savedRune = c
    r.hasSavedRune = true
    r.column--
    if c == '\n' {
        r.lineno--
        r.column = r.lastNewlineColumn
    }
}

func (r *Reader) emitToken(tag tokenTag, val string) {
    r.ch <- scanResult{tok: token{tag, val, r.lineno, r.column}}
}

type scanResult struct {
    tok token
    err error
}

type tokenTag int

const (
    token_EOF tokenTag = iota
    token_LP
    token_RP
    token_DOT
    token_QUOTE
    token_BACKQUOTE
    token_COMMA
    token_COMMAAT
    token_STRING
    token_ATOM
)

type token struct {
    tag    tokenTag
    value  string
    lineno int
    column int
}

func scan(r *Reader) {
    for {
        c, ok := r.readRune()
        if !ok {
            break
        }
        if unicode.IsSpace(c) {
            continue
        }
        switch c {
        case ';':
            skipComment(r)
        case '(':
            r.emitToken(token_LP, "(")
        case ')':
            r.emitToken(token_RP, ")")
        case '.':
            r.emitToken(token_DOT, ".")
        case ',':
            scanPostComma(r)
        case '`':
            r.emitToken(token_BACKQUOTE, "`")
        case '\'':
            r.emitToken(token_QUOTE, "'")
        case '"':
            scanString(r)
        default:
            scanAtom(r, c)
        }
    }
    r.emitToken(token_EOF, "")
    close(r.ch)
}

func skipComment(r *Reader) {
    for {
        c, ok := r.readRune()
        if !ok || c == '\n' {
            break
        }
    }
}

func scanPostComma(r *Reader){
    c, ok := r.readRune()
    if !ok || c != '@' {
        r.unreadRune(c)
        r.emitToken(token_COMMA, ",")
        return
    }
    r.emitToken(token_COMMAAT, ",@")
}

func scanString(r *Reader) {
    begLineno := r.lineno
    begColumn := r.column
    buf := []rune{}

    for {
        c, ok := r.readRune()
        if !ok {
            r.ch <- scanResult{err: fmt.Errorf("%d:%d: unterminated string literal", begLineno, begColumn)}
            return
        }
        switch c {
        case '\\':
            scanEscapeSequence(r, buf)
        case '"':
            r.emitToken(token_STRING, string(buf))
            return
        default:
            buf = append(buf, c)
        }
    }
}

var escMap = map[rune]rune {
    'a': '\a',
    'b': '\b',
    't': '\t',
    'n': '\n',
    'v': '\v',
    'f': '\f',
    'r': '\r',
    '"': '"',
    '\'': '\'',
    '\\': '\\',
}

func scanEscapeSequence(r *Reader, buf []rune) {
    c, ok := r.readRune()
    if !ok {
        return
    }
    if x, ok := escMap[c]; ok {
        buf = append(buf, x)
        return
    }
    buf = append(buf, c)
}

func scanAtom(r *Reader, fc rune) {
    buf := []rune{fc}
    for {
        c, ok := r.readRune()
        if !ok {
            break
        }
        if isAtomDelimiter(c) {
            r.unreadRune(c)
            break
        }
        buf = append(buf, c)
    }
    r.emitToken(token_ATOM, string(buf))
}

var delimTable = map[rune]bool {
    ';': true,
    '(': true,
    ')': true,
    ',': true,
    '`': true,
    '\'': true,
    '"': true,
}

func isAtomDelimiter(c rune) bool {
    if _, ok := delimTable[c]; ok {
        return true
    }
    return unicode.IsSpace(c)
}

