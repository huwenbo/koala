package countlog

import (
	"fmt"
	"encoding/base64"
	"context"
	"reflect"
)

type HumanReadableFormat struct {
	ContextPropertyNames []string
	StringLengthCap int
}

func (hrf *HumanReadableFormat) FormatLog(event Event) string {
	msg := []byte{}
	ctx := hrf.describeContext(event)
	if len(ctx) == 0 {
		msg = append(msg, fmt.Sprintf(
			"=== %s ===\n", event.Event[len("event!"):])...)
	} else {
		msg = append(msg, fmt.Sprintf(
			"=== [%s] %s ===\n", string(ctx), event.Event[len("event!"):])...)
	}
	for i := 0; i < len(event.Properties); i += 2 {
		k, _ := event.Properties[i].(string)
		switch k {
		case "", "ctx", "timestamp":
			continue
		}
		v := event.Properties[i+1]
		formattedV := formatV(v)
		if formattedV == "" {
			continue
		}
		if hrf.StringLengthCap > 0 {
			lenCap := len(formattedV)
			if hrf.StringLengthCap < lenCap {
				lenCap = hrf.StringLengthCap
				formattedV = formattedV[:lenCap] + "...more, capped"
			}
		}
		msg = append(msg, k...)
		msg = append(msg, ": "...)
		msg = append(msg, formattedV...)
		msg = append(msg, '\n')
	}
	return string(msg)
}

func formatV(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	switch typedV := v.(type) {
	case []byte:
		buf := typedV
		if isBinary(buf) {
			return base64.StdEncoding.EncodeToString(buf)
		} else {
			return string(buf)
		}
	case string:
		return typedV
	default:
		stringer, _ := v.(fmt.Stringer)
		if stringer != nil {
			return stringer.String()
		}
		goStringer, _ := v.(fmt.GoStringer)
		if goStringer != nil {
			return goStringer.GoString()
		}
		switch reflect.TypeOf(v).Kind() {
		case reflect.Chan, reflect.Struct, reflect.Interface, reflect.Func, reflect.Array, reflect.Slice, reflect.Ptr:
			return ""
		}
		return fmt.Sprintf("%v", typedV)
	}
}

func (hrf *HumanReadableFormat) describeContext(event Event) []byte {
	msg := []byte{}
	ctx, _ := event.Get("ctx").(context.Context)
	for _, propName := range hrf.ContextPropertyNames {
		propValue := event.Get(propName)
		if propValue == nil && ctx != nil {
			propValue = ctx.Value(propName)
		}
		if propValue != nil {
			if len(msg) > 0 {
				msg = append(msg, ',')
			}
			msg = append(msg, propName...)
			msg = append(msg, '=')
			msg = append(msg, fmt.Sprintf("%v", propValue)...)
		}
	}
	return msg
}

func isBinary(buf []byte) bool {
	for _, b := range buf {
		if b == '\r' || b == '\n' {
			continue
		}
		if b < 32 || b > 127 {
			return true
		}
	}
	return false
}