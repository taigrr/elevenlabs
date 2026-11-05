package encode

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/taigrr/elevenlabs/v2/types"
)

var (
	fileType = reflect.TypeFor[types.File]()
	timeType = reflect.TypeFor[time.Time]()
)

// Query adds the `query`-tagged fields of v (a struct or pointer to one) to
// values.
//
// Encoding rules shared by the query and form encoders, driven by struct
// tags (`query:"name"` / `form:"name"`):
//   - nil pointers, nil slices/maps and empty strings are omitted;
//   - time.Time values are encoded as RFC 3339 in UTC;
//   - slices of scalars are repeated (key=a&key=b);
//   - in forms, maps, structs and slices of structs are JSON-encoded.
func Query(values url.Values, v any) error {
	return walkTagged(v, "query", func(name string, field reflect.Value) error {
		return appendValues(name, field, func(key, val string) { values.Add(key, val) }, false)
	})
}

func walkTagged(v any, tag string, fn func(name string, field reflect.Value) error) error {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("elevenlabs: cannot encode %s as %s", rv.Type(), tag)
	}
	if !rv.CanAddr() {
		addressable := reflect.New(rv.Type()).Elem()
		addressable.Set(rv)
		rv = addressable
	}
	rt := rv.Type()
	for i := range rt.NumField() {
		sf := rt.Field(i)
		name, _, _ := strings.Cut(sf.Tag.Get(tag), ",")
		if name == "" && sf.Anonymous && sf.IsExported() && sf.Type.Kind() == reflect.Struct {
			if err := walkTagged(rv.Field(i).Addr().Interface(), tag, fn); err != nil {
				return err
			}
			continue
		}
		if name == "" || name == "-" || !sf.IsExported() {
			continue
		}
		if err := fn(name, rv.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

// appendValues flattens field into string values. When jsonComposite is
// true, composite values are JSON-encoded instead of rejected.
func appendValues(name string, field reflect.Value, add func(key, val string), jsonComposite bool) error {
	if field.Type() == timeType {
		add(name, field.Interface().(time.Time).UTC().Format(time.RFC3339Nano))
		return nil
	}
	switch field.Kind() {
	case reflect.Pointer, reflect.Interface:
		if field.IsNil() {
			return nil
		}
		return appendValues(name, field.Elem(), add, jsonComposite)
	case reflect.String:
		if field.Len() > 0 {
			add(name, field.String())
		}
		return nil
	case reflect.Bool:
		add(name, strconv.FormatBool(field.Bool()))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		add(name, strconv.FormatInt(field.Int(), 10))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		add(name, strconv.FormatUint(field.Uint(), 10))
		return nil
	case reflect.Float32, reflect.Float64:
		add(name, strconv.FormatFloat(field.Float(), 'f', -1, field.Type().Bits()))
		return nil
	case reflect.Slice, reflect.Array:
		if field.Kind() == reflect.Slice && field.IsNil() {
			return nil
		}
		if isScalarKind(elemKind(field.Type().Elem())) {
			for i := range field.Len() {
				if err := appendValues(name, field.Index(i), add, jsonComposite); err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.Map:
		if field.IsNil() {
			return nil
		}
	}
	if !jsonComposite {
		return fmt.Errorf("elevenlabs: cannot encode field %q of type %s", name, field.Type())
	}
	b, err := json.Marshal(field.Interface())
	if err != nil {
		return fmt.Errorf("elevenlabs: encode field %q: %w", name, err)
	}
	add(name, string(b))
	return nil
}

func elemKind(t reflect.Type) reflect.Kind {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind()
}

func isScalarKind(k reflect.Kind) bool {
	switch k {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// CollectFiles returns every types.File referenced by a form struct, in order.
func CollectFiles(v any) []*types.File {
	var files []*types.File
	_ = walkTagged(v, "form", func(_ string, field reflect.Value) error {
		files = append(files, filesOf(field)...)
		return nil
	})
	return files
}

func filesOf(field reflect.Value) []*types.File {
	switch {
	case field.Type() == fileType:
		if field.CanAddr() {
			return []*types.File{field.Addr().Interface().(*types.File)}
		}
	case field.Kind() == reflect.Pointer && field.Type().Elem() == fileType:
		if !field.IsNil() {
			return []*types.File{field.Interface().(*types.File)}
		}
	case field.Kind() == reflect.Slice && field.Type().Elem() == fileType:
		files := make([]*types.File, field.Len())
		for i := range field.Len() {
			files[i] = field.Index(i).Addr().Interface().(*types.File)
		}
		return files
	}
	return nil
}

// WriteForm writes the multipart representation of v to mw.
func WriteForm(mw *multipart.Writer, v any) error {
	return walkTagged(v, "form", func(name string, field reflect.Value) error {
		if files := filesOf(field); files != nil || isFileField(field.Type()) {
			for _, f := range files {
				if err := writeFilePart(mw, name, f); err != nil {
					return err
				}
			}
			return nil
		}
		var werr error
		err := appendValues(name, field, func(key, val string) {
			if werr == nil {
				werr = mw.WriteField(key, val)
			}
		}, true)
		if err != nil {
			return err
		}
		return werr
	})
}

func isFileField(t reflect.Type) bool {
	return t == fileType ||
		(t.Kind() == reflect.Pointer && t.Elem() == fileType) ||
		(t.Kind() == reflect.Slice && t.Elem() == fileType)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func writeFilePart(mw *multipart.Writer, field string, f *types.File) error {
	if f.Reader == nil {
		return fmt.Errorf("elevenlabs: file field %q has a nil reader", field)
	}
	name := f.Name
	if name == "" {
		name = field
	}
	contentType := f.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		quoteEscaper.Replace(field), quoteEscaper.Replace(name)))
	header.Set("Content-Type", contentType)
	part, err := mw.CreatePart(header)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, f.Reader); err != nil {
		return fmt.Errorf("elevenlabs: upload %q: %w", name, err)
	}
	return nil
}
