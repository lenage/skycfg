// Copyright 2020 The Skycfg Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package protomodule

import (
	"errors"
	"testing"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	pb "github.com/stripe/skycfg/internal/testdata/test_proto"
)

func init() {
	resolve.AllowFloat = true
}

func newRegistry() *protoregistry.Types {
	registry := &protoregistry.Types{}
	registry.RegisterMessage((&pb.MessageV2{}).ProtoReflect().Type())
	registry.RegisterMessage((&pb.MessageV3{}).ProtoReflect().Type())
	registry.RegisterEnum((pb.ToplevelEnumV2)(0).Type())
	registry.RegisterEnum((pb.ToplevelEnumV3)(0).Type())
	return registry
}

func TestProtoPackage(t *testing.T) {
	globals := starlark.StringDict{
		//"proto": NewModule(newRegistry()),
		"proto": &starlarkstruct.Module{
			Name: "proto",
			Members: starlark.StringDict{
				"package": starlarkPackageFn(newRegistry()),
			},
		},
	}

	tests := []struct {
		expr    string
		want    string
		wantErr error
	}{
		{
			expr: `proto.package("skycfg.test_proto")`,
			want: `<proto.Package "skycfg.test_proto">`,
		},
		{
			expr: `dir(proto.package("skycfg.test_proto"))`,
			want: `["MessageV2", "MessageV3", "ToplevelEnumV2", "ToplevelEnumV3"]`,
		},
		{
			expr: `proto.package("skycfg.test_proto").MessageV2`,
			want: `<proto.MessageType "skycfg.test_proto.MessageV2">`,
		},
		{
			expr: `proto.package("skycfg.test_proto").ToplevelEnumV2`,
			want: `<proto.EnumType "skycfg.test_proto.ToplevelEnumV2">`,
		},
		{
			expr:    `proto.package("skycfg.test_proto").NoExist`,
			wantErr: errors.New(`Protobuf type "skycfg.test_proto.NoExist" not found`),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			val, err := starlark.Eval(&starlark.Thread{}, "", test.expr, globals)

			if test.wantErr != nil {
				if !checkError(err, test.wantErr) {
					t.Fatalf("eval(%q): expected error %v, got %v", test.expr, test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("eval(%q): %v", test.expr, err)
			}
			if test.want != val.String() {
				t.Errorf("eval(%q): expected value %q, got %q", test.expr, test.want, val.String())
			}
		})
	}
}

func TestMessageType(t *testing.T) {
	globals := starlark.StringDict{
		"pb": newProtoPackage(newRegistry(), "skycfg.test_proto"),
	}

	tests := []struct {
		expr    string
		want    string
		wantErr error
	}{
		{
			expr: `pb.MessageV2`,
			want: `<proto.MessageType "skycfg.test_proto.MessageV2">`,
		},
		{
			expr: `dir(pb.MessageV2)`,
			want: `["NestedEnum", "NestedMessage"]`,
		},
		{
			expr: `pb.MessageV2.NestedMessage`,
			want: `<proto.MessageType "skycfg.test_proto.MessageV2.NestedMessage">`,
		},
		{
			expr: `pb.MessageV2.NestedEnum`,
			want: `<proto.EnumType "skycfg.test_proto.MessageV2.NestedEnum">`,
		},
		{
			expr:    `pb.MessageV2.NoExist`,
			wantErr: errors.New(`Protobuf type "skycfg.test_proto.MessageV2.NoExist" not found`),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			val, err := starlark.Eval(&starlark.Thread{}, "", test.expr, globals)

			if test.wantErr != nil {
				if !checkError(err, test.wantErr) {
					t.Fatalf("eval(%q): expected error %v, got %v", test.expr, test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("eval(%q): %v", test.expr, err)
			}
			if test.want != val.String() {
				t.Errorf("eval(%q): expected value %q, got %q", test.expr, test.want, val.String())
			}
		})
	}
}

func TestEnumType(t *testing.T) {
	globals := starlark.StringDict{
		"pb": newProtoPackage(newRegistry(), "skycfg.test_proto"),
	}

	tests := []struct {
		expr    string
		want    string
		wantErr error
	}{
		{
			expr: `pb.ToplevelEnumV2`,
			want: `<proto.EnumType "skycfg.test_proto.ToplevelEnumV2">`,
		},
		{
			expr: `dir(pb.ToplevelEnumV2)`,
			want: `["TOPLEVEL_ENUM_V2_A", "TOPLEVEL_ENUM_V2_B"]`,
		},
		{
			expr: `pb.MessageV2.NestedEnum`,
			want: `<proto.EnumType "skycfg.test_proto.MessageV2.NestedEnum">`,
		},
		{
			expr: `dir(pb.MessageV2.NestedEnum)`,
			want: `["NESTED_ENUM_A", "NESTED_ENUM_B"]`,
		},
		{
			expr:    `pb.ToplevelEnumV2.NoExist`,
			wantErr: errors.New(`proto.EnumType has no .NoExist field or method`),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			val, err := starlark.Eval(&starlark.Thread{}, "", test.expr, globals)

			if test.wantErr != nil {
				if !checkError(err, test.wantErr) {
					t.Fatalf("eval(%q): expected error %v, got %v", test.expr, test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("eval(%q): %v", test.expr, err)
			}
			if test.want != val.String() {
				t.Errorf("eval(%q): expected value %q, got %q", test.expr, test.want, val.String())
			}
		})
	}
}

func TestListType(t *testing.T) {
	var listFieldDesc protoreflect.FieldDescriptor
	msg := (&pb.MessageV3{}).ProtoReflect().Descriptor()
	listFieldDesc = msg.Fields().ByName("r_string")

	globals := starlark.StringDict{
		"list": starlark.NewBuiltin("list", func(
			t *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			return newProtoRepeated(listFieldDesc), nil
		}),
	}

	tests := []struct {
		expr    string
		exprFun string
		want    string
		wantErr error
	}{
		{
			expr: `list()`,
			want: `[]`,
		},
		{
			expr: `dir(list())`,
			want: `["append", "clear", "extend", "index", "insert", "pop", "remove"]`,
		},
		// List methods
		{
			exprFun: `
def fun():
    l = list()
    l.append("some string")
    return l
`,
			want: `["some string"]`,
		},
		{
			exprFun: `
def fun():
    l = list()
    l.extend(["a", "b"])
    return l
`,
			want: `["a", "b"]`,
		},
		{
			exprFun: `
def fun():
    l = list()
    l.extend(["a", "b"])
    l.clear()
    return l
`,
			want: `[]`,
		},
		{
			exprFun: `
def fun():
    l = list()
    l.extend(["a", "b"])
    l[1] = "c"
    return l
`,
			want: `["a", "c"]`,
		},

		// List typechecking
		{
			expr:    `list().append(1)`,
			wantErr: errors.New(`TypeError: value 1 (type "int") can't be assigned to type "string".`),
		},
		{
			expr:    `list().extend([1,2])`,
			wantErr: errors.New(`TypeError: value 1 (type "int") can't be assigned to type "string".`),
		},
		{
			exprFun: `
def fun():
    l = list()
    l.extend(["a", "b"])
    l[1] = 1
    return l
`,
			wantErr: errors.New(`TypeError: value 1 (type "int") can't be assigned to type "string".`),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			var val starlark.Value
			var err error
			if test.expr != "" {
				val, err = starlark.Eval(&starlark.Thread{}, "", test.expr, globals)
			} else {
				val, err = evalFunc(test.exprFun, globals)
			}

			if test.wantErr != nil {
				if !checkError(err, test.wantErr) {
					t.Fatalf("eval(%q): expected error %v, got %v", test.expr, test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("eval(%q): %v", test.expr, err)
			}
			if test.want != val.String() {
				t.Errorf("eval(%q): expected value %q, got %q", test.expr, test.want, val.String())
			}
		})
	}
}

func TestMapType(t *testing.T) {
	var mapFieldDesc protoreflect.FieldDescriptor
	msg := (&pb.MessageV3{}).ProtoReflect().Descriptor()
	mapFieldDesc = msg.Fields().ByName("map_string")

	globals := starlark.StringDict{
		"map": starlark.NewBuiltin("map", func(
			t *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			return newProtoMap(mapFieldDesc.MapKey(), mapFieldDesc.MapValue()), nil
		}),
	}

	tests := []struct {
		expr    string
		exprFun string
		want    string
		wantErr error
	}{
		{
			expr: `map()`,
			want: `{}`,
		},
		{
			expr: `dir(map())`,
			want: `["clear", "get", "items", "keys", "pop", "popitem", "setdefault", "update", "values"]`,
		},
		// Map methods
		{
			exprFun: `
def fun():
    m = map()
    m["a"] = "A"
    m.setdefault('a', 'Z')
    m.setdefault('b', 'Z')
    return m
`,
			want: `{"a": "A", "b": "Z"}`,
		},
		{
			exprFun: `
def fun():
    m = map()
    m["a"] = "some string"
    return m
`,
			want: `{"a": "some string"}`,
		},
		{
			exprFun: `
def fun():
    m = map()
    m.update([("a", "a_string"), ("b", "b_string")])
    return m
`,
			want: `{"a": "a_string", "b": "b_string"}`,
		},
		{
			exprFun: `
def fun():
    m = map()
    m["a"] = "some string"
    m.clear()
    return m
`,
			want: `{}`,
		},
		{
			exprFun: `
def fun():
    l = list()
    l.extend(["a", "b"])
    l[1] = "c"
    return l
`,
			want: `["a", "c"]`,
		},

		// Map typechecking
		{
			exprFun: `
def fun():
    m = map()
    m["a"] = 1
    return m
`,
			wantErr: errors.New(`TypeError: value 1 (type "int") can't be assigned to type "string".`),
		},
		{
			expr:    `map().update([("a", 1)])`,
			wantErr: errors.New(`TypeError: value 1 (type "int") can't be assigned to type "string".`),
		},
		{
			expr:    `map().setdefault("a", 1)`,
			wantErr: errors.New(`TypeError: value 1 (type "int") can't be assigned to type "string".`),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			var val starlark.Value
			var err error
			if test.expr != "" {
				val, err = starlark.Eval(&starlark.Thread{}, "", test.expr, globals)
			} else {
				val, err = evalFunc(test.exprFun, globals)
			}

			if test.wantErr != nil {
				if !checkError(err, test.wantErr) {
					t.Fatalf("eval(%q): expected error %v, got %v", test.expr, test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("eval(%q): %v", test.expr, err)
			}
			if test.want != val.String() {
				t.Errorf("eval(%q): expected value %q, got %q", test.expr, test.want, val.String())
			}
		})
	}
}

func evalFunc(src string, globals starlark.StringDict) (starlark.Value, error) {
	globals, err := starlark.ExecFile(&starlark.Thread{}, "", src, globals)
	if err != nil {
		return nil, err
	}
	v, ok := globals["fun"]
	if !ok {
		return nil, errors.New(`Expected function "fun", not found`)
	}
	fun, ok := v.(starlark.Callable)
	if !ok {
		return nil, errors.New("Fun not callable")
	}
	return starlark.Call(&starlark.Thread{}, fun, nil, nil)
}

func checkError(got, want error) bool {
	if got == nil {
		return false
	}
	return got.Error() == want.Error()
}
