// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package messages

import (
	"fmt"
	"time"

	"go.elastic.co/fastjson"
)

// MarshalFastJSON implements the JSON interface for the value type
func (val *Value) MarshalFastJSON(w *fastjson.Writer) error {
	switch typ := val.GetKind().(type) {
	case *Value_NullValue:
		w.RawString("null")
		return nil
	case *Value_Float32Value:
		w.Float32(typ.Float32Value)
	case *Value_Float64Value:
		w.Float64(typ.Float64Value)
		return nil
	case *Value_Int32Value:
		w.Int64(int64(typ.Int32Value))
		return nil
	case *Value_Int64Value:
		w.Int64(typ.Int64Value)
		return nil
	case *Value_Uint32Value:
		w.Uint64(uint64(typ.Uint32Value))
		return nil
	case *Value_Uint64Value:
		w.Uint64(typ.Uint64Value)
		return nil
	case *Value_StringValue:
		w.String(typ.StringValue)
		return nil
	case *Value_BoolValue:
		w.Bool(typ.BoolValue)
		return nil
	case *Value_StructValue:
		err := typ.StructValue.MarshalFastJSON(w)
		if err != nil {
			return fmt.Errorf("error marshaling within value: %w", err)
		}
		// return data, nil
	case *Value_ListValue:
		err := typ.ListValue.MarshalFastJSON(w)
		if err != nil {
			return fmt.Errorf("error marshaling within value: %w", err)
		}
		return nil
	case *Value_TimestampValue:
		w.RawByte('"')
		w.Time(typ.TimestampValue.AsTime(), time.RFC3339Nano)
		w.RawByte('"')
	default:
		return fmt.Errorf("Unknown type %T in event", typ)
	}
	return nil
}

// MarshalFastJSON implements the JSON interface for the struct type
func (sv *Struct) MarshalFastJSON(w *fastjson.Writer) error {
	if sv.GetData() == nil {
		return nil
	}
	w.RawByte('{')
	beginning := true
	for key, val := range sv.GetData() {
		if !beginning {
			w.RawByte(',')
		} else {
			beginning = false
		}

		w.RawString("\"")
		w.RawString(key)
		w.RawString("\":")
		err := val.MarshalFastJSON(w)
		if err != nil {
			return fmt.Errorf("error marshaling value in map: %w", err)
		}
	}
	w.RawByte('}')
	return nil
}

// MarshalFastJSON implements the JSON interface for the list Value type
func (lv *ListValue) MarshalFastJSON(w *fastjson.Writer) error {
	if lv.GetValues() == nil {
		return nil
	}
	w.RawByte('[')
	for iter, val := range lv.GetValues() {
		if iter > 0 {
			w.RawByte(',')
		}
		val.MarshalFastJSON(w)
	}
	w.RawByte(']')
	return nil
}
