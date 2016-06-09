package messages

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/structures"
)

const (
	// NilMarker represents the encoding marker byte for a nil object
	NilMarker = 0xC0

	// TrueMarker represents the encoding marker byte for a true boolean object
	TrueMarker = 0xC3
	// FalseMarker represents the encoding marker byte for a false boolean object
	FalseMarker = 0xC2

	// Int8Marker represents the encoding marker byte for a int8 object
	Int8Marker = 0xC8
	// Int16Marker represents the encoding marker byte for a int16 object
	Int16Marker = 0xC9
	// Int32Marker represents the encoding marker byte for a int32 object
	Int32Marker = 0xCA
	// Int64Marker represents the encoding marker byte for a int64 object
	Int64Marker = 0xCB

	// FloatMarker represents the encoding marker byte for a float32/64 object
	FloatMarker = 0xC1

	// TinyStringMarker represents the encoding marker byte for a string object
	TinyStringMarker = 0x80
	// String8Marker represents the encoding marker byte for a string object
	String8Marker = 0xD0
	// String16Marker represents the encoding marker byte for a string object
	String16Marker = 0xD1
	// String32Marker represents the encoding marker byte for a string object
	String32Marker = 0xD2

	// TinySliceMarker represents the encoding marker byte for a slice object
	TinySliceMarker = 0x90
	// Slice8Marker represents the encoding marker byte for a slice object
	Slice8Marker = 0xD4
	// Slice16Marker represents the encoding marker byte for a slice object
	Slice16Marker = 0xD5
	// Slice32Marker represents the encoding marker byte for a slice object
	Slice32Marker = 0xD6

	// TinyMapMarker represents the encoding marker byte for a map object
	TinyMapMarker = 0xA0
	// Map8Marker represents the encoding marker byte for a map object
	Map8Marker = 0xD8
	// Map16Marker represents the encoding marker byte for a map object
	Map16Marker = 0xD9
	// Map32Marker represents the encoding marker byte for a map object
	Map32Marker = 0xDA

	// TinyStructMarker represents the encoding marker byte for a struct object
	TinyStructMarker = 0xB0
	// Struct8Marker represents the encoding marker byte for a struct object
	Struct8Marker = 0xDC
	// Struct16Marker represents the encoding marker byte for a struct object
	Struct16Marker = 0xDD
)

// Encoder encodes objects of different types to the given stream.
// Attempts to support all builtin golang types, when it can be confidently
// mapped to a data type from: http://alpha.neohq.net/docs/server-manual/bolt-serialization.html#bolt-packstream-structures
// (version v3.1.0-M02 at the time of writing this.
//
// Maps and Slices are a special case, where only
// map[string]interface{} and []interface{} are supported.
// The interface for maps and slices may be more permissive in the future.
type Encoder struct {
	io.Writer
}

// NewEncoder Creates a new Encoder object
func NewEncoder(w io.Writer) Encoder {
	return Encoder{Writer: w}
}

// Encode encodes an object to the stream
func (e Encoder) Encode(iVal interface{}) error {

	// TODO: How to handle pointers?
	//if reflect.TypeOf(iVal) == reflect.Ptr {
	//	return Encode(*iVal)
	//}

	var err error
	switch val := iVal.(type) {
	case nil:
		err = e.encodeNil()
	case bool:
		err = e.encodeBool(val)
	case int:
		err = e.encodeInt(int64(val))
	case int8:
		err = e.encodeInt(int64(val))
	case int16:
		err = e.encodeInt(int64(val))
	case int32:
		err = e.encodeInt(int64(val))
	case int64:
		err = e.encodeInt(int64(val))
	case uint:
		err = e.encodeInt(int64(val))
	case uint8:
		err = e.encodeInt(int64(val))
	case uint16:
		err = e.encodeInt(int64(val))
	case uint32:
		err = e.encodeInt(int64(val))
	case uint64:
		// TODO: Bolt docs only mention going up to int64, not uint64
		// So I'll make this fail for now
		if val > math.MaxInt64 {
			return fmt.Errorf("Integer too big: %d. Max integer supported: %d", val, math.MaxInt64)
		}
		err = e.encodeInt(int64(val))
	case float32:
		err = e.encodeFloat(float64(val))
	case float64:
		err = e.encodeFloat(val)
	case string:
		err = e.encodeString(val)
	case []interface{}:
		// TODO: Support specific slice types?
		err = e.encodeSlice(val)
	case map[string]interface{}:
		// TODO: Support keys other than strings?
		// TODO: Support specific map types?
		err = e.encodeMap(val)
	case structures.Structure:
		err = e.encodeStructure(val)
	default:
		// TODO: How to handle rune or byte?
		return fmt.Errorf("Unrecognized type when encoding data for Bolt transport: %T %+v", val, val)
	}

	return err
}

// encodeNil encodes a nil object to the stream
func (e Encoder) encodeNil() error {
	_, err := e.Write([]byte{NilMarker})
	return err
}

// encodeBool encodes a nil object to the stream
func (e Encoder) encodeBool(val bool) error {
	var err error
	if val {
		_, err = e.Write([]byte{TrueMarker})
	} else {
		_, err = e.Write([]byte{FalseMarker})
	}
	return err
}

// encodeInt encodes a nil object to the stream
func (e Encoder) encodeInt(val int64) error {
	var err error
	switch {
	case val >= -9223372036854775808 && val <= -2147483649:
		// Write as INT_64
		if _, err = e.Write([]byte{Int64Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, val)
	case val >= -2147483648 && val <= -32769:
		// Write as INT_32
		if _, err = e.Write([]byte{Int32Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int32(val))
	case val >= -32768 && val <= -129:
		// Write as INT_16
		if _, err = e.Write([]byte{Int16Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int16(val))
	case val >= -128 && val <= -17:
		// Write as INT_8
		if _, err = e.Write([]byte{Int8Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int8(val))
	case val >= -16 && val <= 127:
		// Write as TINY_INT
		err = binary.Write(e, binary.BigEndian, int8(val))
	case val >= 128 && val <= 32767:
		// Write as INT_16
		if _, err = e.Write([]byte{Int16Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int16(val))
	case val >= 32768 && val <= 2147483647:
		// Write as INT_32
		if _, err = e.Write([]byte{Int32Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, int32(val))
	case val >= 2147483648 && val <= 9223372036854775807:
		// Write as INT_64
		if _, err = e.Write([]byte{Int64Marker}); err != nil {
			return err
		}
		err = binary.Write(e, binary.BigEndian, val)
	default:
		// Can't happen, but if I change the implementation for uint64
		// I want to catch the case if I missed it
		return fmt.Errorf("String too long to write: %d", val)
	}
	return err
}

// encodeFloat encodes a nil object to the stream
func (e Encoder) encodeFloat(val float64) error {
	if _, err := e.Write([]byte{FloatMarker}); err != nil {
		return err
	}
	err := binary.Write(e, binary.BigEndian, val)
	return err
}

// encodeString encodes a nil object to the stream
func (e Encoder) encodeString(val string) error {
	var err error
	bytes := []byte(val)

	length := len(bytes)
	switch {
	case length <= 15:
		if _, err := e.Write([]byte{byte(TinyStringMarker + length)}); err != nil {
			return err
		}
		_, err = e.Write(bytes)
	case length >= 16 && length <= 255:
		if _, err := e.Write([]byte{String8Marker, byte(length)}); err != nil {
			return err
		}
		_, err = e.Write(bytes)
	case length >= 256 && length <= 65535:
		if _, err := e.Write([]byte{String16Marker, byte(length)}); err != nil {
			return err
		}
		_, err = e.Write(bytes)
	case length >= 65536 && length <= 4294967295:
		if _, err := e.Write([]byte{String32Marker, byte(length)}); err != nil {
			// encodeNil encodes a nil object to the stream
			return err
		}
		_, err = e.Write(bytes)
	default:
		// TODO: Can this happen? Does go have a limit on the length?
		// Quick google turned up nothing
		return fmt.Errorf("String too long to write: %s", val)
	}
	return err
}

// encodeSlice encodes a nil object to the stream
func (e Encoder) encodeSlice(val []interface{}) error {
	length := len(val)
	switch {
	case length <= 15:
		if _, err := e.Write([]byte{byte(TinySliceMarker + length)}); err != nil {
			return err
		}
	case length >= 16 && length <= 255:
		if _, err := e.Write([]byte{Slice8Marker, byte(length)}); err != nil {
			return err
		}
	case length >= 256 && length <= 65535:
		if _, err := e.Write([]byte{Slice16Marker, byte(length)}); err != nil {
			return err
		}
	case length >= 65536 && length <= 4294967295:
		if _, err := e.Write([]byte{Slice32Marker, byte(length)}); err != nil {
			return err
		}
	default:
		// TODO: Can this happen? Does go have a limit on the length?
		return fmt.Errorf("Slice too long to write: %+v", val)
	}

	// Encode Slice values
	for _, item := range val {
		if err := e.Encode(item); err != nil {
			return err
		}
	}

	return nil
}

// encodeMap encodes a nil object to the stream
func (e Encoder) encodeMap(val map[string]interface{}) error {
	length := len(val)
	switch {
	case length <= 15:
		if _, err := e.Write([]byte{byte(TinyMapMarker + length)}); err != nil {
			return err
		}
	case length >= 16 && length <= 255:
		if _, err := e.Write([]byte{Map8Marker, byte(length)}); err != nil {
			return err
		}
	case length >= 256 && length <= 65535:
		if _, err := e.Write([]byte{Map16Marker, byte(length)}); err != nil {
			return err
		}
	case length >= 65536 && length <= 4294967295:
		if _, err := e.Write([]byte{Map32Marker, byte(length)}); err != nil {
			return err
		}
	default:
		// TODO: Can this happen? Does go have a limit on the length?
		return fmt.Errorf("Map too long to write: %+v", val)
	}

	// Encode Map values
	for k, v := range val {
		if err := e.Encode(k); err != nil {
			return err
		}
		if err := e.Encode(v); err != nil {
			return err
		}
	}

	return nil
}

// encodeStructure encodes a nil object to the stream
func (e Encoder) encodeStructure(val structures.Structure) error {
	e.Write([]byte{byte(val.Signature())})

	fields := val.Fields()
	length := len(fields)
	switch {
	case length <= 15:
		if _, err := e.Write([]byte{byte(TinyStructMarker + length)}); err != nil {
			return err
		}
	case length >= 16 && length <= 255:
		if _, err := e.Write([]byte{Struct8Marker, byte(length)}); err != nil {
			return err
		}
	case length >= 256 && length <= 65535:
		if _, err := e.Write([]byte{Struct16Marker, byte(length)}); err != nil {
			return err
		}
	default:
		// TODO: Can this happen? Does go have a limit on the length?
		return fmt.Errorf("Structure too long to write: %+v", val)
	}

	for _, field := range fields {
		if err := e.Encode(field); err != nil {
			return err
		}
	}

	return nil
}

