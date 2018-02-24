package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/b71729/opendcm/dictionary"
)

type DicomFile struct {
	filepath       string
	Reader         DicomFileReader
	Preamble       [128]byte
	TotalMetaBytes int64
	Elements       map[uint32]Element
}

// Note for the morning: how to represent elements inside a dicomfile? should elements in sequences be unique? should they be listed at the top level?

type DicomFileChannel struct {
	DicomFile DicomFile
	Error     error
}

// GetElement returns an Element inside the DicomFile according to `tag`.
// If the tag is not found, param `bool` will be false.
func (df DicomFile) GetElement(tag uint32) (Element, bool) {
	e, ok := df.Elements[tag]
	return e, ok
}

// Element represents a data element (see: NEMA 7.1 Data Elements)
type Element struct {
	*dictionary.DictEntry
	ValueLength  uint32
	value        *bytes.Buffer
	LittleEndian bool
	Items        []Item
}

// Item represents a nested Item within a Sequence (see: NEMA 7.5 Nesting of Data Sets)
type Item struct {
	Elements        map[uint32]Element
	UnknownSections [][]byte
}

// GetElement returns an Element inside the DicomFile according to `tag`.
// If the tag is not found, param `bool` will be false.
func (i Item) GetElement(tag uint32) (Element, bool) {
	e, ok := i.Elements[tag]
	return e, ok
}

// LookupTag searches for the corresponding `dictionary.DicomDictionary` entry for the given tag uint32
func LookupTag(t uint32) (*dictionary.DictEntry, bool) {
	val, ok := dictionary.DicomDictionary[t]
	if !ok {
		tag := dictionary.Tag(t)
		name := fmt.Sprintf("Unknown%s", tag)
		return &dictionary.DictEntry{Tag: tag, Name: name, NameHuman: name, VR: "UN", Retired: false}, false
	}
	return val, ok
}

// LookupUID searches for the corresponding `dictionary.UIDDictionary` entry for given uid string
func LookupUID(uid string) (*dictionary.UIDEntry, error) {
	val, ok := dictionary.UIDDictionary[uid]
	if !ok {
		return &dictionary.UIDEntry{}, errors.New("could not find UID")
	}
	return val, nil
}

// ByTag implements a sort interface
type ByTag []Element

func (a ByTag) Len() int           { return len(a) }
func (a ByTag) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTag) Less(i, j int) bool { return a[i].Tag < a[j].Tag }

// Value returns an appropriate representation of the underlying bytestream according to VR
func (i Item) Value() interface{} {
	// TODO
	return nil
}

func RepresentationFromBuffer(buffer *bytes.Buffer, VR string, LittleEndian bool) interface{} {
	switch VR {
	case "UI", "SH", "UT", "ST", "PN", "OW", "LT", "IS", "DS", "CS", "AS", "AE", "LO":
		return string(buffer.Bytes())
	case "UL":
		if LittleEndian {
			return binary.LittleEndian.Uint32(buffer.Bytes())
		}
		return binary.BigEndian.Uint32(buffer.Bytes())
	case "US":
		if LittleEndian {
			return binary.LittleEndian.Uint16(buffer.Bytes())
		}
		return binary.BigEndian.Uint16(buffer.Bytes())
	case "SQ":
		return "asd"
	default:
		return buffer.Bytes()
	}
}

// Value returns an appropriate representation of the underlying bytestream according to VR
func (e Element) Value() interface{} {
	if e.value == nil {
		if len(e.Items) > 0 {
			return e.Items
		} else {
			return nil // neither value nor items set -- contents are empty
		}
	}
	return RepresentationFromBuffer(e.value, e.VR, e.LittleEndian)
}
