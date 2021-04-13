package jsonappender

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func gopterParams() *gopter.TestParameters {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1000
	parameters.SetSeed(1)
	return parameters
}

func TestFloat64(t *testing.T) {
	properties := gopter.NewProperties(gopterParams())
	properties.Property("same as encoding/json", prop.ForAll(
		func(val float64, buf string) bool {
			got, err := Float64(val, []byte(buf))
			return matchesEncodingJSON(val, []byte(buf), got, err)
		}, gen.Float64(), gen.AnyString(),
	))
	properties.TestingRun(t)
}

func TestInt64(t *testing.T) {
	properties := gopter.NewProperties(gopterParams())
	properties.Property("same as encoding/json", prop.ForAll(
		func(val int64, buf string) bool {
			got := Int64(val, []byte(buf))
			return matchesEncodingJSON(val, []byte(buf), got, nil)
		}, gen.Int64(), gen.AnyString(),
	))
	properties.TestingRun(t)
}

func TestUint64(t *testing.T) {
	properties := gopter.NewProperties(gopterParams())
	properties.Property("same as encoding/json", prop.ForAll(
		func(val uint64, buf string) bool {
			got := Uint64(val, []byte(buf))
			return matchesEncodingJSON(val, []byte(buf), got, nil)
		}, gen.UInt64(), gen.AnyString(),
	))
	properties.TestingRun(t)
}

func TestBool(t *testing.T) {
	properties := gopter.NewProperties(gopterParams())
	properties.Property("same as encoding/json", prop.ForAll(
		func(val bool, buf string) bool {
			got := Bool(val, []byte(buf))
			return matchesEncodingJSON(val, []byte(buf), got, nil)
		}, gen.Bool(), gen.AnyString(),
	))
	properties.TestingRun(t)
}

func TestString(t *testing.T) {
	properties := gopter.NewProperties(gopterParams())
	properties.Property("same as encoding/json", prop.ForAll(
		func(val, buf string) bool {
			got := String(val, []byte(buf))
			return matchesEncodingJSON(val, []byte(buf), got, nil)
		}, gen.AnyString(), gen.AnyString(),
	))
	properties.TestingRun(t)
}

func TestTime(t *testing.T) {
	properties := gopter.NewProperties(gopterParams())
	properties.Property("same as encoding/json", prop.ForAll(
		func(val time.Time, buf string) bool {
			got, err := Time(val, []byte(buf))
			return matchesEncodingJSON(val, []byte(buf), got, err)
		}, gen.Time(), gen.AnyString(),
	))
	properties.TestingRun(t)
}

func TestFieldName(t *testing.T) {
	properties := gopter.NewProperties(gopterParams())
	properties.Property("same as encoding/json", prop.ForAll(
		func(val, buf string) bool {
			got := FieldName(val, []byte(buf))
			if got[len(got)-1] != ':' {
				return false
			}
			got = got[:len(got)-1]
			return matchesEncodingJSON(val, []byte(buf), got, nil)
		}, gen.AnyString(), gen.AnyString(),
	))
	properties.TestingRun(t)
}

func matchesEncodingJSON(val interface{}, buf, got []byte, gotErr error) bool {
	want, wantErr := encodingJSONAppend(val, buf)
	if wantErr != nil {
		return gotErr != nil
	}
	if gotErr != nil {
		return false
	}
	want = bytes.TrimSuffix(want, []byte{'\n'})
	return string(want) == string(got)
}

func encodingJSONAppend(val interface{}, bts []byte) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(true)
	err := enc.Encode(&val)
	if err != nil {
		return nil, err
	}
	return append(bts, buf.Bytes()...), nil
}
