package parser

/*
	https://github.com/pacedotdev/oto.git
	The MIT License (MIT)
	Copyright (c) 2021 Pace Software Ltd

	Modifications:
	Removed oto specific fields.
*/

import "github.com/pkg/errors"

// ErrNotFound is returned when an Object is not found.
var ErrNotFound = errors.New("not found")

// Definition describes an definition.
type Definition struct {
	// PackageName is the name of the package.
	PackageName string `json:"packageName"`
	// Services are the services described in this definition.
	Services []Service `json:"services"`
	// Objects are the structures that are used throughout this definition.
	Objects []Object `json:"objects"`
	// Imports is a map of Go imports that should be imported into
	// Go code.
	Imports map[string]string `json:"imports"`
}

// Object describes a data structure that is part of this definition.
type Object struct {
	TypeID   string  `json:"typeID"`
	Name     string  `json:"name"`
	Imported bool    `json:"imported"`
	Fields   []Field `json:"fields"`
	Comment  string  `json:"comment"`
}

// Object looks up an object by name. Returns ErrNotFound error
// if it cannot find it.
func (d *Definition) Object(name string) (*Object, error) {
	for i := range d.Objects {
		obj := &d.Objects[i]
		if obj.Name == name {
			return obj, nil
		}
	}
	return nil, ErrNotFound
}

// ObjectIsInput gets whether this object is a method
// input (request) type or not.
// Returns true if any method.InputObject.ObjectName matches
// name.
func (d *Definition) ObjectIsInput(name string) bool {
	for _, service := range d.Services {
		for _, method := range service.Methods {
			for i := range method.InputObjects {

				if method.InputObjects[i].ObjectName == name {
					return true
				}
			}
		}
	}
	return false
}

// ObjectIsOutput gets whether this object is a method
// output (response) type or not.
// Returns true if any method.OutputObject.ObjectName matches
// name.
func (d *Definition) ObjectIsOutput(name string) bool {
	for _, service := range d.Services {
		for _, method := range service.Methods {
			for i := range method.OutputObjects {

				if method.OutputObjects[i].ObjectName == name {
					return true
				}
			}
		}
	}
	return false
}

type Service struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
	Comment string   `json:"comment"`
}

// Method describes a method that a Service can perform.
type Method struct {
	Name          string      `json:"name"`
	InputObjects  []FieldType `json:"inputObjects"`
	OutputObjects []FieldType `json:"outputObjects"`
	Comment       string      `json:"comment"`
}

// Field describes the field inside an Object.
type Field struct {
	Name       string              `json:"name"`
	Type       FieldType           `json:"type"`
	Comment    string              `json:"comment"`
	Tag        string              `json:"tag"`
	ParsedTags map[string]FieldTag `json:"parsedTags"`
}

// FieldTag is a parsed tag.
// For more information, see Struct Tags in Go.
type FieldTag struct {
	// Value is the value of the tag.
	Value string `json:"value"`
	// Options are the options for the tag.
	Options []string `json:"options"`
}

// FieldType holds information about the type of data that this
// Field stores.
type FieldType struct {
	TypeID     string `json:"typeID"`
	TypeName   string `json:"typeName"`
	ObjectName string `json:"objectName"`
	// CleanObjectName is the ObjectName with * removed
	// for pointer types.
	CleanObjectName string `json:"cleanObjectName"`
	Multiple        bool   `json:"multiple"`
	Package         string `json:"package"`
	IsObject        bool   `json:"isObject"`
}
