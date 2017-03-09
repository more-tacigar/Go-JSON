package json

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
)

type Encoder struct {
	writer io.Writer
	buf    bytes.Buffer
}

// NewEncoder creates and initializes a new Encoder using w as its initiali writer.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
	}
}

func (e *Encoder) Encode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if err := e.encValue(rv); err != nil {
		return err
	}
	e.writer.Write(e.buf.Bytes())
	return nil
}

func (e *Encoder) encValue(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if _, err := e.buf.WriteString(strconv.FormatInt(rv.Int(), 10)); err != nil {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if _, err := e.buf.WriteString(strconv.FormatFloat(rv.Float(), 'f', 10, 64)); err != nil {
			return err
		}
	case reflect.Bool:
		if _, err := e.buf.WriteString(strconv.FormatBool(rv.Bool())); err != nil {
			return err
		}
	case reflect.String:
		return e.encStringLiteral(rv)
	case reflect.Slice, reflect.Array:
		return e.encArrayOrSlice(rv)
	case reflect.Map:
		return e.encMap(rv)
	case reflect.Struct:
		return e.encStruct(rv)
	default:
		return nil //uso
	}
	return nil
}

func (e *Encoder) encStringLiteral(rv reflect.Value) error {
	if _, err := e.buf.WriteRune('"'); err != nil {
		return err
	}
	if _, err := e.buf.WriteString(rv.String()); err != nil {
		return err
	}
	if _, err := e.buf.WriteRune('"'); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encArrayOrSlice(rv reflect.Value) error {
	if _, err := e.buf.WriteRune('['); err != nil {
		return err
	}
	for i := 0; i < rv.Len(); i++ {
		if err := e.encValue(rv.Index(i)); err != nil {
			return err
		}
		if i < rv.Len()-1 {
			if _, err := e.buf.WriteRune(','); err != nil {
				return err
			}
		}
	}
	if _, err := e.buf.WriteRune(']'); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encMap(rv reflect.Value) error {
	if _, err := e.buf.WriteRune('{'); err != nil {
		return err
	}
	for i, key := range rv.MapKeys() {
		if key.Kind() != reflect.String {
			return nil // uso
		}
		if _, err := e.buf.WriteRune('"'); err != nil {
			return err
		}
		if _, err := e.buf.WriteString(key.String()); err != nil {
			return err
		}
		if _, err := e.buf.WriteString(`":`); err != nil {
			return err
		}
		if err := e.encValue(rv.MapIndex(key)); err != nil {
			return err
		}
		if i < rv.Len()-1 {
			if _, err := e.buf.WriteRune(','); err != nil {
				return err
			}
		}
	}
	if _, err := e.buf.WriteRune('}'); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encStruct(rv reflect.Value) error {
	if _, err := e.buf.WriteRune('{'); err != nil {
		return err
	}
	for i := 0; i < rv.Type().NumField(); i++ {
		key := rv.Type().Field(i).Name
		if _, err := e.buf.WriteRune('"'); err != nil {
			return err
		}
		if _, err := e.buf.WriteString(key); err != nil {
			return err
		}
		if _, err := e.buf.WriteString(`":`); err != nil {
			return err
		}
		if err := e.encValue(rv.Field(i)); err != nil {
			return err
		}
		if i < rv.NumField()-1 {
			if _, err := e.buf.WriteRune(','); err != nil {
				return err
			}
		}
	}
	if _, err := e.buf.WriteRune('}'); err != nil {
		return err
	}
	return nil
}
