// Code generated by serde. DO NOT EDIT.

//go:build durable

package coroutine

import astutil "golang.org/x/tools/go/ast/astutil"
import reflect "reflect"
import unicode "unicode"
import constraint "go/build/constraint"
import exec "os/exec"
import ast "go/ast"
import scanner "go/scanner"
import rand "math/rand"
import regexp "regexp"
import bytes "bytes"
import big "math/big"
import syntax "regexp/syntax"
import types "go/types"
import os "os"
import runtime "runtime"
import strconv "strconv"
import base64 "encoding/base64"
import build "go/build"
import comment "go/doc/comment"
import constant "go/constant"
import parser "go/parser"
import log "log"
import bufio "bufio"
import strings "strings"
import syscall "syscall"
import io "io"
import objectpath "golang.org/x/tools/go/types/objectpath"
import atomic "sync/atomic"
import sort "sort"
import json "encoding/json"
import crypto "crypto"
import unsafe "unsafe"
import packages "golang.org/x/tools/go/packages"
import token "go/token"
import sync "sync"
import slog "log/slog"
import serde "github.com/stealthrocket/coroutine/serde"
import fs "io/fs"
import doc "go/doc"
import typeutil "golang.org/x/tools/go/types/typeutil"
import time "time"
import semver "golang.org/x/mod/semver"

func Serialize_gen8(s *serde.Serializer, x []any, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	b = serde.SerializeSize(len(x), b)
	for _, x := range x {
		b = serde.SerializeInterface(s, x, b)
	}
	return b
}

func Deserialize_gen8(d *serde.Deserializer, b []byte) ([]any, []byte) {
	d = serde.EnsureDeserializer(d)
	n, b := serde.DeserializeSize(b)
	var z []any
	for i := 0; i < n; i++ {
		var x any
		x, b = serde.DeserializeInterface(d, b)
		z = append(z, x)
	}
	return z, b
}

func Serialize_gen7(s *serde.Serializer, x struct{ objects []any }, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	{
		x := x.objects
		b = Serialize_gen8(s, x, b)
	}
	return b
}

func Deserialize_gen7(d *serde.Deserializer, b []byte) (struct{ objects []any }, []byte) {
	d = serde.EnsureDeserializer(d)
	var z struct{ objects []any }
	{
		var x []any
		x, b = Deserialize_gen8(d, b)
		z.objects = x
	}
	return z, b
}

func Serialize_Storage(s *serde.Serializer, z Storage, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	x := (struct{ objects []any })(z)
	b = Serialize_gen7(s, x, b)
	return b
}
func Deserialize_Storage(d *serde.Deserializer, b []byte) (Storage, []byte) {
	d = serde.EnsureDeserializer(d)
	var x struct{ objects []any }
	x, b = Deserialize_gen7(d, b)
	return (Storage)(x), b
}
func Serialize_gen5(s *serde.Serializer, x struct {
	IP int
	Storage
	Resume bool
}, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	{
		x := x.IP
		b = serde.SerializeInt(s, x, b)
	}
	{
		x := x.Storage
		b = Serialize_Storage(s, x, b)
	}
	{
		x := x.Resume
		b = serde.SerializeBool(s, x, b)
	}
	return b
}

func Deserialize_gen5(d *serde.Deserializer, b []byte) (struct {
	IP int
	Storage
	Resume bool
}, []byte) {
	d = serde.EnsureDeserializer(d)
	var z struct {
		IP int
		Storage
		Resume bool
	}
	{
		var x int
		x, b = serde.DeserializeInt(d, b)
		z.IP = x
	}
	{
		var x Storage
		x, b = Deserialize_Storage(d, b)
		z.Storage = x
	}
	{
		var x bool
		x, b = serde.DeserializeBool(d, b)
		z.Resume = x
	}
	return z, b
}

func Serialize_Frame(s *serde.Serializer, z Frame, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	x := (struct {
		IP int
		Storage
		Resume bool
	})(z)
	b = Serialize_gen5(s, x, b)
	return b
}
func Deserialize_Frame(d *serde.Deserializer, b []byte) (Frame, []byte) {
	d = serde.EnsureDeserializer(d)
	var x struct {
		IP int
		Storage
		Resume bool
	}
	x, b = Deserialize_gen5(d, b)
	return (Frame)(x), b
}
func Serialize_gen3(s *serde.Serializer, x []Frame, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	b = serde.SerializeSize(len(x), b)
	for _, x := range x {
		b = Serialize_Frame(s, x, b)
	}
	return b
}

func Deserialize_gen3(d *serde.Deserializer, b []byte) ([]Frame, []byte) {
	d = serde.EnsureDeserializer(d)
	n, b := serde.DeserializeSize(b)
	var z []Frame
	for i := 0; i < n; i++ {
		var x Frame
		x, b = Deserialize_Frame(d, b)
		z = append(z, x)
	}
	return z, b
}

func Serialize_gen1(s *serde.Serializer, x struct {
	FP     int
	Frames []Frame
}, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	{
		x := x.FP
		b = serde.SerializeInt(s, x, b)
	}
	{
		x := x.Frames
		b = Serialize_gen3(s, x, b)
	}
	return b
}

func Deserialize_gen1(d *serde.Deserializer, b []byte) (struct {
	FP     int
	Frames []Frame
}, []byte) {
	d = serde.EnsureDeserializer(d)
	var z struct {
		FP     int
		Frames []Frame
	}
	{
		var x int
		x, b = serde.DeserializeInt(d, b)
		z.FP = x
	}
	{
		var x []Frame
		x, b = Deserialize_gen3(d, b)
		z.Frames = x
	}
	return z, b
}

func Serialize_Stack(s *serde.Serializer, z Stack, b []byte) []byte {
	s = serde.EnsureSerializer(s)
	x := (struct {
		FP     int
		Frames []Frame
	})(z)
	b = Serialize_gen1(s, x, b)
	return b
}
func Deserialize_Stack(d *serde.Deserializer, b []byte) (Stack, []byte) {
	d = serde.EnsureDeserializer(d)
	var x struct {
		FP     int
		Frames []Frame
	}
	x, b = Deserialize_gen1(d, b)
	return (Stack)(x), b
}
func init() {
	var t reflect.Type
	{
		var x uint16
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x time.Timer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.SelectDir
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.FuncLit
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.IndexExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x rand.Zipf
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.SyntaxError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x build.NoGoError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Named
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x int8
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.TypeName
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.BasicLit
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x build.Package
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Termios
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.SelectCase
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x parser.Mode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x build.Context
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x crypto.Hash
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x unsafe.Pointer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.Pinner
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RtAttr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.ChanDir
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.Regexp
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.DocLink
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.NetlinkRouteAttr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.CommClause
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.FuncDecl
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x big.Accuracy
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.Record
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.TextHandler
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x byte
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x unicode.RangeTable
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.BasicKind
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Checker
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x serde.Generator
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x sort.StringSlice
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.TypeList
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.EmptyStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ChanType
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.Package
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x complex128
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.NetlinkRouteRequest
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x io.OffsetWriter
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.Kind
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.TypeAndValue
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.TypeSpec
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x bufio.Reader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x build.ImportMode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.Delim
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x rune
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x atomic.Uint64
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RawSockaddrInet6
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x time.ParseError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x bufio.Scanner
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.List
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.Source
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x bool
		t = reflect.TypeOf(x)
		sw := func(s *serde.Serializer, x any, b []byte) []byte {
			return serde.SerializeBool(s, x.(bool), b)
		}
		dw := func(d *serde.Deserializer, b []byte) (any, []byte) {
			return serde.DeserializeBool(d, b)
		}
		serde.RegisterTypeWithCodec(t, sw, dw)
	}
	{
		var x sync.Pool
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Initializer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.Kind
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x strings.Builder
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x strconv.NumError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x big.ErrNaN
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.LevelVar
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x time.Location
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.TypeParamList
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.Scope
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x constraint.OrExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x exec.Error
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.Inst
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.Prog
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.HandlerOptions
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x typeutil.Hasher
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.Value
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x int32
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Inet6Pktinfo
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.NlAttr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Ustat_t
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Msghdr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x io.SectionReader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.IndexListExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.DeclStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x scanner.Mode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.InstOp
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x uintptr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x atomic.Int64
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.PtraceRegs
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Time_t
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.TypeParam
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.ErrorCode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Italic
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x os.LinkError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.InotifyEvent
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x token.FileSet
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.Module
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x typeutil.MethodSetCache
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x uint8
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x atomic.Value
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x unicode.CaseRange
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.SelectorExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x token.Token
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Builtin
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.GoStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.LoadMode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.InvalidUTF8Error
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Slice
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x sync.Map
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Tuple
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.SwitchStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x bufio.Writer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.Number
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.MarshalerError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.UnmarshalTypeError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Credential
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x io.PipeWriter
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.TypeAssertExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Basic
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x uint64
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x sync.WaitGroup
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x sync.Cond
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Package
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.CompositeLit
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x exec.ExitError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Parser
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x serde.Serializer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x serde.ID
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RawSockaddrAny
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RtNexthop
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Tms
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Instance
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.BinaryExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.RangeStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.Config
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x log.Logger
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x doc.Type
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Struct
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x os.ProcessState
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.SliceHeader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Link
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RawSockaddrInet4
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x time.Weekday
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Config
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Context
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x os.Process
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Errno
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.NlMsghdr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x unicode.Range32
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Scope
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x big.Int
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x bufio.ReadWriter
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x int
		t = reflect.TypeOf(x)
		sw := func(s *serde.Serializer, x any, b []byte) []byte {
			return serde.SerializeInt(s, x.(int), b)
		}
		dw := func(d *serde.Deserializer, b []byte) (any, []byte) {
			return serde.DeserializeInt(d, b)
		}
		serde.RegisterTypeWithCodec(t, sw, dw)
	}
	{
		var x time.Ticker
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.MergeMode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.BadExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x regexp.Regexp
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SysProcAttr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.PanicNilError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Utsname
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x time.Month
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.Method
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x rand.Rand
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x constraint.TagExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x doc.Package
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x token.Position
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x typeutil.Map
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Fsid
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SockaddrUnix
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Union
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.BadDecl
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x constraint.AndExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x constraint.SyntaxError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.Decoder
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.RawMessage
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.Frames
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.StackRecord
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Timespec
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Statfs_t
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x doc.Func
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.ProcAttr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SysProcIDMap
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Dirent
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.IfAddrmsg
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.Ellipsis
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.BadStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x base64.Encoding
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SockaddrLinklayer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.SelectStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x scanner.Scanner
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x serde.Typedef
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x atomic.Int32
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.GenDecl
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x bytes.Reader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x big.RoundingMode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.Op
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x atomic.Uintptr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Array
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.ChanDir
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.Comment
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ExprStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x fs.PathError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Iovec
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Selection
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x exec.Cmd
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.Flags
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x uint
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.WaitStatus
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.InterfaceType
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.CommentMap
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x constant.Kind
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.InvalidUnmarshalError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Plain
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Linger
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.IfInfomsg
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x objectpath.Path
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Heading
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x Frame
		t = reflect.TypeOf(x)
		sw := func(s *serde.Serializer, x any, b []byte) []byte {
			return Serialize_Frame(s, x.(Frame), b)
		}
		dw := func(d *serde.Deserializer, b []byte) (any, []byte) {
			return Deserialize_Frame(d, b)
		}
		serde.RegisterTypeWithCodec(t, sw, dw)
	}
	{
		var x sort.IntSlice
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SockaddrInet6
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Ucred
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x token.Pos
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.CaseClause
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.SelectionKind
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.CallExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.SendStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.UnsupportedValueError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.Error
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x sync.Once
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.FdSet
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Func
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.StdSizes
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.EpollEvent
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x strings.Reader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Const
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.ImportMode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ChanDir
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.IfStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x scanner.Error
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x big.Rat
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x objectpath.Encoder
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.ListItem
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x complex64
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x atomic.Bool
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x build.MultiplePackageError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.Ident
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.BlockProfileRecord
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RawSockaddr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RawSockaddrUnix
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SocketControlMessage
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.FieldList
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x os.SyscallError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x bytes.Buffer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x fs.FileMode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x io.PipeReader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.MethodSet
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.UnaryExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x constraint.NotExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x uint32
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x sync.RWMutex
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.IPMreq
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.Error
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x Stack
		t = reflect.TypeOf(x)
		sw := func(s *serde.Serializer, x any, b []byte) []byte {
			return Serialize_Stack(s, x.(Stack), b)
		}
		dw := func(d *serde.Deserializer, b []byte) (any, []byte) {
			return Deserialize_Stack(d, b)
		}
		serde.RegisterTypeWithCodec(t, sw, dw)
	}
	{
		var x string
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Stat_t
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.TypeAssertionError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.MemProfileRecord
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SockaddrInet4
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x strings.Replacer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Label
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.PkgName
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.StarExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x doc.Example
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Var
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Pointer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Rlimit
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x time.Duration
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.File
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x os.ProcAttr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RtGenmsg
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SockaddrNetlink
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Utimbuf
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.Object
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x serde.Deserializer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.FuncType
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.ArgumentError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.ModuleError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.Encoder
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.UnsupportedTypeError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Flock_t
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x unicode.SpecialCase
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.AssignStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x semver.ByVersion
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.JSONHandler
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Map
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x os.File
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.Frame
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RawSockaddrNetlink
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ImportSpec
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.SliceExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.ErrorKind
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x astutil.Cursor
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x io.LimitedReader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.NetlinkMessage
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RawSockaddrLinklayer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Chan
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Info
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x build.Directive
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Rusage
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SockFprog
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.TCPInfo
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x token.File
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ParenExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Doc
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x Storage
		t = reflect.TypeOf(x)
		sw := func(s *serde.Serializer, x any, b []byte) []byte {
			return Serialize_Storage(s, x.(Storage), b)
		}
		dw := func(d *serde.Deserializer, b []byte) (any, []byte) {
			return Deserialize_Storage(d, b)
		}
		serde.RegisterTypeWithCodec(t, sw, dw)
	}
	{
		var x sort.Float64Slice
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Timex
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.StructTag
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.LabeledStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.TypeSwitchStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ReturnStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.Package
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x big.Float
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x big.Word
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.Value
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x float64
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Term
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Nil
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ArrayType
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ForStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x doc.Value
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x doc.Mode
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Error
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ObjKind
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syntax.EmptyOp
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.Logger
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Timeval
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Sysinfo_t
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.BlockStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x base64.CorruptInputError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.Func
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x packages.OverlayJSON
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x int16
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Cmsghdr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.SockFilter
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.Field
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.BranchStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.DeferStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.MapType
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x float32
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x atomic.Uint32
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.IPv6MTUInfo
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.IPv6Mreq
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.MapIter
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.ValueError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Paragraph
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.LinkDef
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Interface
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.StructField
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x time.Time
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.ValueSpec
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.BasicInfo
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.CommentGroup
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Code
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.Attr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Signal
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x runtime.MemStats
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.ICMPv6Filter
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.RtMsg
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.NlMsgerr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x reflect.StringHeader
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x types.Signature
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.IncDecStmt
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x json.UnmarshalFieldError
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x slog.Level
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x int64
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x sync.Mutex
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.IPMreqn
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x syscall.Inet4Pktinfo
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x unicode.Range16
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.KeyValueExpr
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x ast.StructType
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x scanner.ErrorList
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x comment.Printer
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
	{
		var x doc.Note
		t = reflect.TypeOf(x)
		serde.RegisterType(t)
	}
}
