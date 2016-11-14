// Package tlv provides a toolkit for working with TLV (Type-Length-Value) objects and TLV object List.
package tlv

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"io"
)

// TLV represents a Type-Length-Value object.
type TLV interface {
	Type() byte
	Length() int32
	Value() []byte
}

type object struct {
	typ byte
	len int32
	val []byte
}

// Type returns the object's type
func (o *object) Type() byte {
	return o.typ
}

// Length returns the object's type
func (o *object) Length() int32 {
	return o.len
}

// Type returns the object's value
func (o *object) Value() []byte {
	return o.val
}

// Equal returns true if a pair of TLV objects are the same.
func Equal(tlv1, tlv2 TLV) bool {
	if tlv1 == nil {
		return tlv2 == nil
	} else if tlv2 == nil {
		return false
	} else if tlv1.Type() != tlv2.Type() {
		return false
	} else if tlv1.Length() != tlv2.Length() {
		return false
	} else if !bytes.Equal(tlv1.Value(), tlv2.Value()) {
		return false
	}
	return true
}

var (
	// ErrTLVRead is returned when there is an error reading a TLV object.
	ErrTLVRead = fmt.Errorf("TLV %s", "read error")
	// ErrTLVWrite is returned when  there is an error writing a TLV object.
	ErrTLVWrite = fmt.Errorf("TLV %s", "write error")
	// ErrTypeNotFound is returned when a request for a TLV type is made and none can be found.
	ErrTypeNotFound = fmt.Errorf("TLV %s", "type not found")
)

// New returns a TLV object from the args
func New(typ byte, val []byte) TLV {
	tlv := new(object)
	tlv.typ = typ
	tlv.len = int32(len(val))
	tlv.val = make([]byte, tlv.Length())
	copy(tlv.val, val)
	return tlv
}

// FromBytes returns a TLV object from bytes
func FromBytes(data []byte) (TLV, error) {
	objBuf := bytes.NewBuffer(data)
	return ReadObject(objBuf)
}

// ToBytes returns bytes from a TLV object
func ToBytes(tlv TLV) ([]byte, error) {
	data := make([]byte, 0)
	objBuf := bytes.NewBuffer(data)
	err := WriteObject(tlv, objBuf)
	return objBuf.Bytes(), err
}

// ReadObject returns a TLV object from io.Reader
func ReadObject(r io.Reader) (TLV, error) {
	tlv := new(object)

	var typ byte
	var err error
	err = binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}
	tlv.typ = typ

	var length int32
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	tlv.len = length

	tlv.val = make([]byte, tlv.Length())
	l, err := r.Read(tlv.val)
	if err != nil {
		return nil, err
	} else if int32(l) != tlv.Length() {
		return tlv, ErrTLVRead
	}

	return tlv, nil
}

// WriteObject writes a TLV object to io.Writer
func WriteObject(tlv TLV, w io.Writer) error {
	var err error

	typ := tlv.Type()
	err = binary.Write(w, binary.BigEndian, typ)
	if err != nil {
		return err
	}

	length := tlv.Length()
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return err
	}

	n, err := w.Write(tlv.Value())
	if err != nil {
		return err
	} else if int32(n) != tlv.Length() {
		return ErrTLVWrite
	}

	return nil
}

// List is ad double-linked list containing TLV objects.
type List struct {
	objects *list.List
}

// NewList returns a new, empty TLVList.
func NewList() *List {
	tl := new(List)
	tl.objects = list.New()
	return tl
}

// Length returns the number of objects int the TLVList.
func (tl *List) Length() int32 {
	return int32(tl.objects.Len())
}

// Get checks the TLVList for any object matching the type, It returns the first one found.
// If the type could not be found, Get returns ErrTypeNotFound.
func (tl *List) Get(typ byte) (TLV, error) {
	for e := tl.objects.Front(); e != nil; e = e.Next() {
		if e.Value.(*object).Type() == typ {
			return e.Value.(*object), nil
		}
	}
	return nil, ErrTypeNotFound
}

// GetAll checks the TLVList for all objects matching the type, returning a slice containing all matching objects.
// If no object has the requested type, an empty slice is returned.
func (tl *List) GetAll(typ byte) []TLV {
	ts := make([]TLV, 0)
	for e := tl.objects.Front(); e != nil; e = e.Next() {
		if e.Value.(*object).Type() == typ {
			ts = append(ts, e.Value.(TLV))
		}
	}
	return ts
}

// Remove removes all objects with the requested type.
// It returns a count of the number of removed objects.
func (tl *List) Remove(typ byte) int {
	var totalRemoved int
	for {
		var removed int
		for e := tl.objects.Front(); e != nil; e = e.Next() {
			if e.Value.(*object).Type() == typ {
				tl.objects.Remove(e)
				removed++
				break
			}
		}
		if removed == 0 {
			break
		}
		totalRemoved += removed
	}
	return totalRemoved
}

// RemoveObject takes an TLV object as an argument, and removes all matching objects.
// It matches on not just type, but also the value contained in the object.
func (tl *List) RemoveObject(obj TLV) int {
	var totalRemoved int
	for {
		var removed int
		for e := tl.objects.Front(); e != nil; e = e.Next() {
			if Equal(e.Value.(*object), obj) {
				tl.objects.Remove(e)
				removed++
				break
			}
		}

		if removed == 0 {
			break
		}
		totalRemoved += removed
	}
	return totalRemoved
}

// Add pushes a new TLV object onto the TLVList. It builds the object from its args
func (tl *List) Add(typ byte, value []byte) {
	obj := New(typ, value)
	tl.objects.PushBack(obj)
}

// AddObject adds a TLV object onto the TLVList
func (tl *List) AddObject(obj TLV) {
	tl.objects.PushBack(obj)
}

// Write writes out the TLVList to an io.Writer.
func (tl *List) Write(w io.Writer) error {
	for e := tl.objects.Front(); e != nil; e = e.Next() {
		err := WriteObject(e.Value.(TLV), w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Read takes an io.Reader and builds a TLVList from that.
func Read(r io.Reader) (*List, error) {
	tl := NewList()
	var err error
	for {
		var tlv TLV
		if tlv, err = ReadObject(r); err != nil {
			break
		}
		tl.objects.PushBack(tlv)
	}

	if err == io.EOF {
		err = nil
	}
	return tl, err
}
