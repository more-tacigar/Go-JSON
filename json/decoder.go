package json

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type (
	Decoder struct {
		lexer *Lexer
	}

	DecoderSyntaxError struct {
		line int
		err  error
	}

	DecoderTypeError struct {
		line     int
		tp       reflect.Kind
		corrects []reflect.Kind
	}
)

var (
	ErrUnsettableValue = errors.New("Unsettable value")
)

// NewDecoder creates and initialize a new Decoder using r as its initial reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		lexer: NewLexer(r),
	}
}

func (d *Decoder) genSyntaxError(e error) error {
	return &DecoderSyntaxError{
		line: d.lexer.currline,
		err:  e,
	}
}

func (e *DecoderSyntaxError) Error() string {
	return fmt.Sprintf("json.Decoder: line %d, %v", e.line, e.err.Error())
}

func (d *Decoder) genTypeError(tp reflect.Kind, corrects []reflect.Kind) error {
	return &DecoderTypeError{
		line:     d.lexer.currline,
		tp:       tp,
		corrects: corrects,
	}
}

func (e *DecoderTypeError) Error() string {
	buf := new(bytes.Buffer)
	for i := 0; i < len(e.corrects); i++ {
		buf.WriteString(e.corrects[i].String())
		if i != len(e.corrects)-1 {
			buf.WriteString("or")
		}
	}
	return fmt.Sprintf("json.Decoder: line %d, %v but want %v", e.line, e.tp, buf.String())
}

// Decode decodes JSON into v.
func (d *Decoder) Decode(v interface{}) error {
	if _, err := d.lexer.Lex(); err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return d.genSyntaxError(ErrUnsettableValue)
	}

	var tmp reflect.Value
	switch rv.Elem().Kind() {
	case reflect.Map:
		tmp = reflect.MakeMap(rv.Type().Elem())
		if err := d.value(tmp); err != nil {
			return err
		}
		rv.Elem().Set(tmp)
	default:
		tmp = reflect.New(reflect.ValueOf(v).Elem().Type())
		if err := d.value(tmp.Elem()); err != nil {
			return err
		}
		rv.Elem().Set(tmp.Elem())
	}
	return nil
}

func (d *Decoder) value(rv reflect.Value) error {
	switch d.lexer.currtoken {
	case TK_LBRACE:
		if err := d.object(rv); err != nil {
			return err
		}
	case TK_LBRACKET:
		return nil
	case TK_TRUE, TK_FALSE:
		var b = false
		if d.lexer.currtoken == TK_TRUE {
			b = true
		}
		if err := d.boolLiteral(rv, b); err != nil {
			return err
		}
		if _, err := d.lexer.Lex(); err != nil {
			return err
		}
		return nil
	case TK_NULL:
		if err := d.nullLiteral(rv); err != nil {
			return err
		}
		if _, err := d.lexer.Lex(); err != nil {
			return err
		}
	case TK_STRING_LITERAL:
		sv := d.lexer.svalue
		if err := d.stringLiteral(rv, sv); err != nil {
			return err
		}
		if _, err := d.lexer.Lex(); err != nil {
			return err
		}
	case TK_NUMBER_LITERAL:
		nv := d.lexer.nvalue
		if err := d.numberLiteral(rv, nv); err != nil {
			return err
		}
		if _, err := d.lexer.Lex(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) object(rv reflect.Value) error {
	if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Map {
		return d.genTypeError(rv.Kind(), []reflect.Kind{reflect.Struct, reflect.Map})
	}
	for {
		// Read a string token and store the string literal as key.
		if _, err := d.lexer.Lex(); err != nil {
			return err
		}
		if d.lexer.currtoken != TK_STRING_LITERAL {
			return d.genSyntaxError(errors.New("This must be a string literal but"))
		}
		key := d.lexer.svalue

		// Read a colon token and skip it.
		if _, err := d.lexer.Lex(); err != nil {
			return err
		}
		if d.lexer.currtoken != TK_COLON {
			return d.genSyntaxError(errors.New("This must be a colon token"))
		}

		// Read a value and store it into the map or struct.
		if _, err := d.lexer.Lex(); err != nil {
			return err
		}
		switch rv.Kind() {
		case reflect.Struct:
			for i := 0; i < rv.Type().NumField(); i++ {
				if rv.Type().Field(i).Name == key {
					if err := d.value(rv.Field(i)); err != nil {
						return err
					}
					goto end_field
				}
			}
		case reflect.Map:
			value := reflect.New(rv.Type().Elem()).Elem()
			if err := d.value(value); err != nil {
				return err
			}
			rv.SetMapIndex(reflect.ValueOf(key), value)
			goto end_field
		}
	end_field:
		if d.lexer.currtoken != TK_COMMA {
			break
		}
	}
	if d.lexer.currtoken != TK_RBRACE {
		return d.genSyntaxError(errors.New("This must be a right brace token"))
	}
	return nil
}

func (d *Decoder) nullLiteral(rv reflect.Value) error {
	if !rv.CanSet() {
		return ErrUnsettableValue
	}
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		rv.Set(reflect.Zero(rv.Type()))
	default:
	}
	return nil
}

func (d *Decoder) boolLiteral(rv reflect.Value, b bool) error {
	if !rv.CanSet() || rv.Kind() != reflect.Bool {
		return ErrUnsettableValue
	}
	rv.SetBool(b)
	return nil
}

func (d *Decoder) stringLiteral(rv reflect.Value, s string) error {
	if !rv.CanSet() || rv.Kind() != reflect.String {
		return ErrUnsettableValue
	}
	rv.SetString(s)
	return nil
}

func (d *Decoder) numberLiteral(rv reflect.Value, n float64) error {
	if !rv.CanSet() {
		return ErrUnsettableValue
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv.SetInt(int64(n))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		rv.SetFloat(n)
	default:
		return ErrUnsettableValue
	}
	return nil
}
