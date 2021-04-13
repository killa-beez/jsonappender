package jsonappender

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"
	"unicode/utf8"
)

// JSONAppender will append json
type JSONAppender interface {
	AppendJSON(buf []byte) ([]byte, error)
}

// BufWriter write json to your writer in a buffered manner. Don't forget to Flush.
// Errors are collected in Error so you don't have to check after each write.
type BufWriter struct {
	Error     error
	writer    *bufio.Writer
	stringBuf []byte
}

// NewBufWriter does what the name says
func NewBufWriter(w io.Writer) *BufWriter {
	bw := BufWriter{
		writer: bufio.NewWriter(w),
	}
	return &bw
}

// Flush flushes the buffer
func (bw *BufWriter) Flush() error {
	if bw.Error != nil {
		return bw.Error
	}
	bw.Error = bw.writer.Flush()
	return bw.Error
}

// Reset resets BufWriter to start writing anew.
func (bw *BufWriter) Reset(w io.Writer) {
	bw.Error = nil
	if bw.writer == nil {
		bw.writer = bufio.NewWriter(w)
		return
	}
	bw.writer.Reset(w)
}

// Raw writes a raw value.
func (bw *BufWriter) Raw(val []byte) {
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.Write(val)
}

// RawString is like Raw but takes a string.
func (bw *BufWriter) RawString(val string) {
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.WriteString(val)
}

// RawByte writes one single byte.
func (bw *BufWriter) RawByte(val byte) {
	if bw.Error != nil {
		return
	}
	bw.Error = bw.writer.WriteByte(val)
}

// Int64 writes an int64 value
func (bw *BufWriter) Int64(val int64) {
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.WriteString(strconv.FormatInt(val, 10))
}

// Int64 append an int64 value
func Int64(val int64, buf []byte) []byte {
	return append(buf, strconv.FormatInt(val, 10)...)
}

// Uint64 writes a uint64 value
func (bw *BufWriter) Uint64(val uint64) {
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.WriteString(strconv.FormatUint(val, 10))
}

// Uint64 append a uint64 value
func Uint64(val uint64, buf []byte) []byte {
	return append(buf, strconv.FormatUint(val, 10)...)
}

// FieldName writes a fieldname in the format: "name":
func (bw *BufWriter) FieldName(name string) {
	if bw.Error != nil {
		return
	}
	bw.stringBuf = String(name, bw.stringBuf[:0])
	bw.stringBuf = append(bw.stringBuf, ':')
	_, bw.Error = bw.writer.Write(bw.stringBuf)
}

// FieldName append a fieldname in the format: "name":
func FieldName(name string, buf []byte) []byte {
	buf = String(name, buf)
	return append(buf, ':')
}

// Bool writes a bool value
func (bw *BufWriter) Bool(val bool) {
	if bw.Error != nil {
		return
	}
	if val {
		_, bw.Error = bw.writer.WriteString("true")
		return
	}
	_, bw.Error = bw.writer.WriteString("false")
}

// Bool append a bool value
func Bool(val bool, buf []byte) []byte {
	if val {
		return append(buf, `true`...)
	}
	return append(buf, `false`...)
}

// Time writes a time.Time value
func (bw *BufWriter) Time(t time.Time) {
	if bw.Error != nil {
		return
	}
	bw.stringBuf, bw.Error = Time(t, bw.stringBuf[:0])
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.Write(bw.stringBuf)
}

// Time append a time.Time value
func Time(t time.Time, buf []byte) ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, fmt.Errorf("Time.MarshalJSON: year outside of range [0,9999]")
	}
	buf = append(buf, '"')
	buf = t.AppendFormat(buf, time.RFC3339Nano)
	buf = append(buf, '"')
	return buf, nil
}

// Float64 writes a float64 value
func (bw *BufWriter) Float64(f float64) {
	if bw.Error != nil {
		return
	}
	bw.stringBuf, bw.Error = Float64(f, bw.stringBuf[:0])
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.Write(bw.stringBuf)
}

// Float64 append a float64 value
func Float64(f float64, buf []byte) ([]byte, error) {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return buf, fmt.Errorf("unsupported value: %s", strconv.FormatFloat(f, 'g', -1, 64))
	}
	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.

	abs := math.Abs(f)
	format := byte('f')
	if abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		format = 'e'
	}
	start := len(buf)
	buf = strconv.AppendFloat(buf, f, format, -1, 64)

	if format == 'e' {
		// clean up e-09 to e-9
		n := len(buf) - start
		if n >= 4 && buf[start+n-4] == 'e' && buf[start+n-3] == '-' && buf[start+n-2] == '0' {
			buf[start+n-2] = buf[start+n-1]
			buf = append(buf[:start], buf[start:start+n-1]...)
		}
	}
	return buf, nil
}

// Value writes any json marshallable value
func (bw *BufWriter) Value(val interface{}) {
	if bw.Error != nil {
		return
	}
	bw.stringBuf, bw.Error = Value(val, bw.stringBuf[:0])
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.Write(bw.stringBuf)
}

// Value appends any json marshallable value
func Value(val interface{}, buf []byte) ([]byte, error) {
	switch v := val.(type) {
	case string:
		return String(v, buf), nil
	case float64:
		return Float64(v, buf)
	case int64:
		return Int64(v, buf), nil
	case int:
		return Int64(int64(v), buf), nil
	case uint64:
		return Uint64(v, buf), nil
	case uint:
		return Uint64(uint64(v), buf), nil
	case time.Time:
		return Time(v, buf)
	case map[string]interface{}:
		return Object(v, buf)
	case []interface{}:
		return Array(v, buf)
	case JSONAppender:
		return v.AppendJSON(buf)
	case json.Marshaler:
		bb, err := v.MarshalJSON()
		return append(buf, bb...), err
	}
	bb, err := json.Marshal(val)
	return append(buf, bb...), err
}

// Object writes an object value
func (bw *BufWriter) Object(mp map[string]interface{}) {
	if bw.Error != nil {
		return
	}
	bw.stringBuf, bw.Error = Object(mp, bw.stringBuf[:0])
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.Write(bw.stringBuf)
}

// Object appends an object value
func Object(mp map[string]interface{}, buf []byte) ([]byte, error) {
	var comma bool
	buf = append(buf, '{')
	var err error
	for k, v := range mp {
		if comma {
			buf = append(buf, ',')
		}
		comma = true
		buf = FieldName(k, buf)
		buf, err = Value(v, buf)
		if err != nil {
			return buf, err
		}

	}
	return append(buf, '}'), nil
}

// Array writes an array value
func (bw *BufWriter) Array(slice []interface{}) {
	if bw.Error != nil {
		return
	}
	bw.stringBuf, bw.Error = Array(slice, bw.stringBuf[:0])
	if bw.Error != nil {
		return
	}
	_, bw.Error = bw.writer.Write(bw.stringBuf)
}

// Array appends an array value
func Array(slice []interface{}, buf []byte) ([]byte, error) {
	var comma bool
	buf = append(buf, '[')
	var err error
	for i := 0; i < len(slice); i++ {
		if comma {
			buf = append(buf, ',')
		}
		comma = true
		buf, err = Value(slice[i], buf)
		if err != nil {
			return buf, err
		}
	}
	return append(buf, ']'), nil
}

// String writes a string value
func (bw *BufWriter) String(val string) {
	if bw.Error != nil {
		return
	}
	bw.stringBuf = String(val, bw.stringBuf[:0])
	_, bw.Error = bw.writer.Write(bw.stringBuf)
}

// String appends a string value
func String(s string, buf []byte) []byte {
	const hex = "0123456789abcdef"
	buf = append(buf, '"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] {
				i++
				continue
			}
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			buf = append(buf, '\\')
			switch b {
			case '\\', '"':
				buf = append(buf, b)
			case '\n':
				buf = append(buf, 'n')
			case '\r':
				buf = append(buf, 'r')
			case '\t':
				buf = append(buf, 't')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// It also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				buf = append(buf, 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			buf = append(buf, '\\', 'u', 'f', 'f', 'd')
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			buf = append(buf, '\\', 'u', '2', '0', '2', hex[c&0xF])
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		buf = append(buf, s[start:]...)
	}
	buf = append(buf, '"')
	return buf
}

// htmlSafeSet holds the value true if the ASCII character with the given
// array position can be safely represented inside a JSON string, embedded
// inside of HTML <script> tags, without any additional escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), the backslash character ("\"), HTML opening and closing
// tags ("<" and ">"), and the ampersand ("&").
var htmlSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      false,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      false,
	'=':      true,
	'>':      false,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
