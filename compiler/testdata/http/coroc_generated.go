// Code generated by coroc. DO NOT EDIT

package main

import (
	"github.com/stealthrocket/coroutine"
	unsafe "unsafe"
	unicode "unicode"
	time "time"
	syscall "syscall"
	sync "sync"
	strings "strings"
	strconv "strconv"
	sort "sort"
	serde "github.com/stealthrocket/coroutine/serde"
	runtime "runtime"
	reflect "reflect"
	os "os"
	net "net"
	multipart "mime/multipart"
	quotedprintable "mime/quotedprintable"
	mime "mime"
	log "log"
	list "container/list"
	io "io"
	http "net/http"
	url "net/url"
	textproto "net/textproto"
	netip "net/netip"
	httptrace "net/http/httptrace"
	internal "net/http/internal"
	fs "io/fs"
	flate "compress/flate"
	gzip "compress/gzip"
	embed "embed"
	crypto "crypto"
	crc32 "hash/crc32"
	cgo "runtime/cgo"
	bytes "bytes"
	bufio "bufio"
	big "math/big"
	rand "math/rand"
	atomic "sync/atomic"
	asn1_1 "encoding/asn1"
	pem "encoding/pem"
	hex "encoding/hex"
	base64 "encoding/base64"
	asn1 "vendor/golang.org/x/crypto/cryptobyte/asn1"
	dnsmessage "vendor/golang.org/x/net/dns/dnsmessage"
	route "vendor/golang.org/x/net/route"
	idna "vendor/golang.org/x/net/idna"
	httpproxy "vendor/golang.org/x/net/http/httpproxy"
	hpack "vendor/golang.org/x/net/http2/hpack"
	cryptobyte "vendor/golang.org/x/crypto/cryptobyte"
	chacha20 "vendor/golang.org/x/crypto/chacha20"
	bidi "vendor/golang.org/x/text/unicode/bidi"
	transform "vendor/golang.org/x/text/transform"
	norm "vendor/golang.org/x/text/unicode/norm"
	bidirule "vendor/golang.org/x/text/secure/bidirule"
	aes "crypto/aes"
	x509 "crypto/x509"
	tls "crypto/tls"
	rsa "crypto/rsa"
	rc4 "crypto/rc4"
	pkix "crypto/x509/pkix"
	elliptic "crypto/elliptic"
	ed25519 "crypto/ed25519"
	ecdsa "crypto/ecdsa"
	ecdh "crypto/ecdh"
	dsa "crypto/dsa"
	des "crypto/des"
	cipher "crypto/cipher"
)

func RoundTrip(req *http.Request) (_ *http.Response, _ error) {
	_c := coroutine.LoadContext[*http.Request, *http.Response]()
	_f, _fp := _c.Push()
	var _o0 *http.Response
	if _f.IP > 0 {
		if _v := _f.Get(0); _v != nil {
			req = _v.(*http.Request)
		}
		if _v := _f.Get(1); _v != nil {
			_o0 = _v.(*http.Response)
		}
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, req)
			_f.Set(1, _o0)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = coroutine.Yield[*http.Request, *http.Response](req)
		_f.IP = 2
		fallthrough
	case _f.IP < 3:
		return _o0, nil
	}
	return
}

func work() {
	_c := coroutine.LoadContext[*http.Request, *http.Response]()
	_f, _fp := _c.Push()
	var _o0 *http.Response
	var _o1 error
	var _o2 bool
	if _f.IP > 0 {
		if _v := _f.Get(0); _v != nil {
			_o0 = _v.(*http.Response)
		}
		if _v := _f.Get(1); _v != nil {
			_o1 = _v.(error)
		}
		if _v := _f.Get(2); _v != nil {
			_o2 = _v.(bool)
		}
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _o0)
			_f.Set(1, _o1)
			_f.Set(2, _o2)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0, _o1 = http.Get("http://example.com")
		_f.IP = 2
		fallthrough
	case _f.IP < 4:
		switch {
		case _f.IP < 3:
			_o2 = _o1 != nil
			_f.IP = 3
			fallthrough
		case _f.IP < 4:
			if _o2 {
				panic(_o1)
			}
		}
		_f.IP = 4
		fallthrough
	case _f.IP < 5:

		fmt.Println(_o0.StatusCode)
	}
}

func main() {
	http.DefaultTransport = &yieldingRoundTripper{}

	c := coroutine.New[*http.Request, *http.Response](work)

	for c.Next() {
		req := c.Recv()
		fmt.Println("Requesting", req.URL.String())
		c.Send(&http.Response{
			StatusCode: 200,
		})
	}
}
func init() {
	serde.RegisterType[**byte]()
	serde.RegisterType[*[100000]uintptr]()
	serde.RegisterType[*[131072]uint16]()
	serde.RegisterType[*[133]byte]()
	serde.RegisterType[*[140737488355327]byte]()
	serde.RegisterType[*[16]byte]()
	serde.RegisterType[*[16]int32]()
	serde.RegisterType[*[171]uint8]()
	serde.RegisterType[*[19]int]()
	serde.RegisterType[*[1]uintptr]()
	serde.RegisterType[*[256]byte]()
	serde.RegisterType[*[28]byte]()
	serde.RegisterType[*[28]uint8]()
	serde.RegisterType[*[29]byte]()
	serde.RegisterType[*[2]byte]()
	serde.RegisterType[*[2]float32]()
	serde.RegisterType[*[2]float64]()
	serde.RegisterType[*[2]int32]()
	serde.RegisterType[*[2]uint32]()
	serde.RegisterType[*[2]uint64]()
	serde.RegisterType[*[2]uintptr]()
	serde.RegisterType[*[316]int]()
	serde.RegisterType[*[32]byte]()
	serde.RegisterType[*[32]rune]()
	serde.RegisterType[*[32]uint8]()
	serde.RegisterType[*[32]uintptr]()
	serde.RegisterType[*[33]byte]()
	serde.RegisterType[*[3]uint32]()
	serde.RegisterType[*[3]uint64]()
	serde.RegisterType[*[48]byte]()
	serde.RegisterType[*[48]uint8]()
	serde.RegisterType[*[49]byte]()
	serde.RegisterType[*[4]byte]()
	serde.RegisterType[*[4]uint64]()
	serde.RegisterType[*[512]uintptr]()
	serde.RegisterType[*[57]byte]()
	serde.RegisterType[*[5]float64]()
	serde.RegisterType[*[65536]uintptr]()
	serde.RegisterType[*[65]byte]()
	serde.RegisterType[*[66]byte]()
	serde.RegisterType[*[66]uint8]()
	serde.RegisterType[*[67]byte]()
	serde.RegisterType[*[6]float64]()
	serde.RegisterType[*[6]uint64]()
	serde.RegisterType[*[70368744177663]uint16]()
	serde.RegisterType[*[8]byte]()
	serde.RegisterType[*[8]uint32]()
	serde.RegisterType[*[8]uint8]()
	serde.RegisterType[*[97]byte]()
	serde.RegisterType[*[9]uint64]()
	serde.RegisterType[*[][]byte]()
	serde.RegisterType[*[]byte]()
	serde.RegisterType[*[]uint64]()
	serde.RegisterType[*bool]()
	serde.RegisterType[*byte]()
	serde.RegisterType[*int]()
	serde.RegisterType[*int32]()
	serde.RegisterType[*int64]()
	serde.RegisterType[*int8]()
	serde.RegisterType[*string]()
	serde.RegisterType[*struct{cstr *byte}]()
	serde.RegisterType[*uint]()
	serde.RegisterType[*uint16]()
	serde.RegisterType[*uint32]()
	serde.RegisterType[*uint64]()
	serde.RegisterType[*uint8]()
	serde.RegisterType[*uintptr]()
	serde.RegisterType[[0]uintptr]()
	serde.RegisterType[[1000]uintptr]()
	serde.RegisterType[[100]byte]()
	serde.RegisterType[[1024]bool]()
	serde.RegisterType[[1024]byte]()
	serde.RegisterType[[1024]int8]()
	serde.RegisterType[[1024]uint8]()
	serde.RegisterType[[1048576]uint8]()
	serde.RegisterType[[104]byte]()
	serde.RegisterType[[104]int8]()
	serde.RegisterType[[107]string]()
	serde.RegisterType[[108]byte]()
	serde.RegisterType[[10]byte]()
	serde.RegisterType[[10]float64]()
	serde.RegisterType[[10]string]()
	serde.RegisterType[[11008]uint64]()
	serde.RegisterType[[11]float64]()
	serde.RegisterType[[126]bool]()
	serde.RegisterType[[127]bool]()
	serde.RegisterType[[128][4]uint64]()
	serde.RegisterType[[128]byte]()
	serde.RegisterType[[128]float32]()
	serde.RegisterType[[128]uint16]()
	serde.RegisterType[[128]uint32]()
	serde.RegisterType[[128]uint64]()
	serde.RegisterType[[128]uintptr]()
	serde.RegisterType[[129]uint8]()
	serde.RegisterType[[12]byte]()
	serde.RegisterType[[12]float64]()
	serde.RegisterType[[12]int8]()
	serde.RegisterType[[131072]uint32]()
	serde.RegisterType[[131072]uintptr]()
	serde.RegisterType[[133]byte]()
	serde.RegisterType[[13]byte]()
	serde.RegisterType[[13]int32]()
	serde.RegisterType[[1408]uint16]()
	serde.RegisterType[[1408]uint8]()
	serde.RegisterType[[144]byte]()
	serde.RegisterType[[14]byte]()
	serde.RegisterType[[14]int8]()
	serde.RegisterType[[15]float64]()
	serde.RegisterType[[16384]byte]()
	serde.RegisterType[[16384]uint8]()
	serde.RegisterType[[16576]uint8]()
	serde.RegisterType[[1664]uint16]()
	serde.RegisterType[[16][16]int32]()
	serde.RegisterType[[16]byte]()
	serde.RegisterType[[16]int]()
	serde.RegisterType[[16]int8]()
	serde.RegisterType[[16]uint32]()
	serde.RegisterType[[16]uint64]()
	serde.RegisterType[[16]uint8]()
	serde.RegisterType[[16]uintptr]()
	serde.RegisterType[[17]int32]()
	serde.RegisterType[[17]string]()
	serde.RegisterType[[18]byte]()
	serde.RegisterType[[19426]byte]()
	serde.RegisterType[[19]int]()
	serde.RegisterType[[19]int32]()
	serde.RegisterType[[1]byte]()
	serde.RegisterType[[1]struct{}]()
	serde.RegisterType[[1]uint32]()
	serde.RegisterType[[1]uint64]()
	serde.RegisterType[[1]uint8]()
	serde.RegisterType[[1]uintptr]()
	serde.RegisterType[[20]byte]()
	serde.RegisterType[[20]uint64]()
	serde.RegisterType[[20]uint8]()
	serde.RegisterType[[21]byte]()
	serde.RegisterType[[248]byte]()
	serde.RegisterType[[2496]uint16]()
	serde.RegisterType[[249]uint8]()
	serde.RegisterType[[24]byte]()
	serde.RegisterType[[252]uintptr]()
	serde.RegisterType[[253]uintptr]()
	serde.RegisterType[[255]byte]()
	serde.RegisterType[[256][]byte]()
	serde.RegisterType[[256]bool]()
	serde.RegisterType[[256]byte]()
	serde.RegisterType[[256]float32]()
	serde.RegisterType[[256]int]()
	serde.RegisterType[[256]int8]()
	serde.RegisterType[[256]uint32]()
	serde.RegisterType[[256]uint64]()
	serde.RegisterType[[256]uint8]()
	serde.RegisterType[[257]uint32]()
	serde.RegisterType[[25]byte]()
	serde.RegisterType[[288]int]()
	serde.RegisterType[[28]byte]()
	serde.RegisterType[[28]uint64]()
	serde.RegisterType[[29]byte]()
	serde.RegisterType[[29]uint64]()
	serde.RegisterType[[2]byte]()
	serde.RegisterType[[2]int]()
	serde.RegisterType[[2]int32]()
	serde.RegisterType[[2]int64]()
	serde.RegisterType[[2]uint16]()
	serde.RegisterType[[2]uint32]()
	serde.RegisterType[[2]uint64]()
	serde.RegisterType[[2]uintptr]()
	serde.RegisterType[[3072]uint16]()
	serde.RegisterType[[30]byte]()
	serde.RegisterType[[32768]uint32]()
	serde.RegisterType[[32]byte]()
	serde.RegisterType[[32]float64]()
	serde.RegisterType[[32]int32]()
	serde.RegisterType[[32]string]()
	serde.RegisterType[[32]uint32]()
	serde.RegisterType[[32]uintptr]()
	serde.RegisterType[[33]byte]()
	serde.RegisterType[[33]float64]()
	serde.RegisterType[[3]byte]()
	serde.RegisterType[[3]float64]()
	serde.RegisterType[[3]uint32]()
	serde.RegisterType[[3]uint64]()
	serde.RegisterType[[4096]byte]()
	serde.RegisterType[[40]byte]()
	serde.RegisterType[[40]int8]()
	serde.RegisterType[[48]byte]()
	serde.RegisterType[[49]byte]()
	serde.RegisterType[[4]byte]()
	serde.RegisterType[[4]float64]()
	serde.RegisterType[[4]string]()
	serde.RegisterType[[4]uint32]()
	serde.RegisterType[[4]uint64]()
	serde.RegisterType[[4]uint8]()
	serde.RegisterType[[4]uintptr]()
	serde.RegisterType[[50]byte]()
	serde.RegisterType[[50]uintptr]()
	serde.RegisterType[[512]byte]()
	serde.RegisterType[[512]uint32]()
	serde.RegisterType[[512]uintptr]()
	serde.RegisterType[[56]byte]()
	serde.RegisterType[[56]int8]()
	serde.RegisterType[[56]uint8]()
	serde.RegisterType[[57]byte]()
	serde.RegisterType[[5]byte]()
	serde.RegisterType[[5]float64]()
	serde.RegisterType[[5]string]()
	serde.RegisterType[[5]uint]()
	serde.RegisterType[[5]uint32]()
	serde.RegisterType[[5]uint64]()
	serde.RegisterType[[5]uint8]()
	serde.RegisterType[[5]uintptr]()
	serde.RegisterType[[607]int64]()
	serde.RegisterType[[61]struct{Size uint32; Mallocs uint64; Frees uint64}]()
	serde.RegisterType[[6208]uint16]()
	serde.RegisterType[[64488]byte]()
	serde.RegisterType[[64]byte]()
	serde.RegisterType[[64]int8]()
	serde.RegisterType[[64]uint32]()
	serde.RegisterType[[64]uint64]()
	serde.RegisterType[[64]uintptr]()
	serde.RegisterType[[65528]byte]()
	serde.RegisterType[[65]byte]()
	serde.RegisterType[[66]byte]()
	serde.RegisterType[[67]byte]()
	serde.RegisterType[[68]byte]()
	serde.RegisterType[[68]struct{Size uint32; Mallocs uint64; Frees uint64}]()
	serde.RegisterType[[68]uint16]()
	serde.RegisterType[[68]uint32]()
	serde.RegisterType[[68]uint64]()
	serde.RegisterType[[68]uint8]()
	serde.RegisterType[[696][2]uint64]()
	serde.RegisterType[[69]uintptr]()
	serde.RegisterType[[6]byte]()
	serde.RegisterType[[6]float64]()
	serde.RegisterType[[6]int]()
	serde.RegisterType[[6]int32]()
	serde.RegisterType[[6]uint64]()
	serde.RegisterType[[6]uint8]()
	serde.RegisterType[[6]uintptr]()
	serde.RegisterType[[72]byte]()
	serde.RegisterType[[768]byte]()
	serde.RegisterType[[78]byte]()
	serde.RegisterType[[7]float64]()
	serde.RegisterType[[7]uint64]()
	serde.RegisterType[[7]uint8]()
	serde.RegisterType[[800]byte]()
	serde.RegisterType[[80]uint64]()
	serde.RegisterType[[8640]uint16]()
	serde.RegisterType[[88]byte]()
	serde.RegisterType[[8][4][16]uint8]()
	serde.RegisterType[[8][64]uint32]()
	serde.RegisterType[[8]byte]()
	serde.RegisterType[[8]float64]()
	serde.RegisterType[[8]int8]()
	serde.RegisterType[[8]string]()
	serde.RegisterType[[8]uint32]()
	serde.RegisterType[[8]uint64]()
	serde.RegisterType[[8]uint8]()
	serde.RegisterType[[92]int8]()
	serde.RegisterType[[96]byte]()
	serde.RegisterType[[97]byte]()
	serde.RegisterType[[9]byte]()
	serde.RegisterType[[9]string]()
	serde.RegisterType[[9]uint64]()
	serde.RegisterType[[]*byte]()
	serde.RegisterType[[][2]uint16]()
	serde.RegisterType[[][32]byte]()
	serde.RegisterType[[][4096]byte]()
	serde.RegisterType[[][]byte]()
	serde.RegisterType[[][]int]()
	serde.RegisterType[[][]int32]()
	serde.RegisterType[[][]rune]()
	serde.RegisterType[[][]uint32]()
	serde.RegisterType[[]byte]()
	serde.RegisterType[[]float32]()
	serde.RegisterType[[]float64]()
	serde.RegisterType[[]int]()
	serde.RegisterType[[]int16]()
	serde.RegisterType[[]int32]()
	serde.RegisterType[[]int64]()
	serde.RegisterType[[]int8]()
	serde.RegisterType[[]rune]()
	serde.RegisterType[[]string]()
	serde.RegisterType[[]uint]()
	serde.RegisterType[[]uint16]()
	serde.RegisterType[[]uint32]()
	serde.RegisterType[[]uint64]()
	serde.RegisterType[[]uint8]()
	serde.RegisterType[[]uintptr]()
	serde.RegisterType[aes.KeySizeError]()
	serde.RegisterType[asn1.Tag]()
	serde.RegisterType[asn1_1.BitString]()
	serde.RegisterType[asn1_1.Enumerated]()
	serde.RegisterType[asn1_1.Flag]()
	serde.RegisterType[asn1_1.ObjectIdentifier]()
	serde.RegisterType[asn1_1.RawContent]()
	serde.RegisterType[asn1_1.RawValue]()
	serde.RegisterType[asn1_1.StructuralError]()
	serde.RegisterType[asn1_1.SyntaxError]()
	serde.RegisterType[atomic.Bool]()
	serde.RegisterType[atomic.Int32]()
	serde.RegisterType[atomic.Int64]()
	serde.RegisterType[atomic.Uint32]()
	serde.RegisterType[atomic.Uint64]()
	serde.RegisterType[atomic.Uintptr]()
	serde.RegisterType[atomic.Value]()
	serde.RegisterType[base64.CorruptInputError]()
	serde.RegisterType[base64.Encoding]()
	serde.RegisterType[bidi.Class]()
	serde.RegisterType[bidi.Direction]()
	serde.RegisterType[bidi.Ordering]()
	serde.RegisterType[bidi.Paragraph]()
	serde.RegisterType[bidi.Properties]()
	serde.RegisterType[bidi.Run]()
	serde.RegisterType[bidirule.Transformer]()
	serde.RegisterType[big.Accuracy]()
	serde.RegisterType[big.ErrNaN]()
	serde.RegisterType[big.Float]()
	serde.RegisterType[big.Int]()
	serde.RegisterType[big.Rat]()
	serde.RegisterType[big.RoundingMode]()
	serde.RegisterType[big.Word]()
	serde.RegisterType[bool]()
	serde.RegisterType[bufio.ReadWriter]()
	serde.RegisterType[bufio.Reader]()
	serde.RegisterType[bufio.Scanner]()
	serde.RegisterType[bufio.Writer]()
	serde.RegisterType[byte]()
	serde.RegisterType[bytes.Buffer]()
	serde.RegisterType[bytes.Reader]()
	serde.RegisterType[cgo.Handle]()
	serde.RegisterType[cgo.Incomplete]()
	serde.RegisterType[chacha20.Cipher]()
	serde.RegisterType[cipher.StreamReader]()
	serde.RegisterType[cipher.StreamWriter]()
	serde.RegisterType[complex128]()
	serde.RegisterType[crc32.Table]()
	serde.RegisterType[crypto.Hash]()
	serde.RegisterType[cryptobyte.BuildError]()
	serde.RegisterType[cryptobyte.Builder]()
	serde.RegisterType[cryptobyte.String]()
	serde.RegisterType[des.KeySizeError]()
	serde.RegisterType[dnsmessage.AAAAResource]()
	serde.RegisterType[dnsmessage.AResource]()
	serde.RegisterType[dnsmessage.Builder]()
	serde.RegisterType[dnsmessage.CNAMEResource]()
	serde.RegisterType[dnsmessage.Class]()
	serde.RegisterType[dnsmessage.Header]()
	serde.RegisterType[dnsmessage.MXResource]()
	serde.RegisterType[dnsmessage.Message]()
	serde.RegisterType[dnsmessage.NSResource]()
	serde.RegisterType[dnsmessage.Name]()
	serde.RegisterType[dnsmessage.OPTResource]()
	serde.RegisterType[dnsmessage.OpCode]()
	serde.RegisterType[dnsmessage.Option]()
	serde.RegisterType[dnsmessage.PTRResource]()
	serde.RegisterType[dnsmessage.Parser]()
	serde.RegisterType[dnsmessage.Question]()
	serde.RegisterType[dnsmessage.RCode]()
	serde.RegisterType[dnsmessage.Resource]()
	serde.RegisterType[dnsmessage.ResourceHeader]()
	serde.RegisterType[dnsmessage.SOAResource]()
	serde.RegisterType[dnsmessage.SRVResource]()
	serde.RegisterType[dnsmessage.TXTResource]()
	serde.RegisterType[dnsmessage.Type]()
	serde.RegisterType[dnsmessage.UnknownResource]()
	serde.RegisterType[dsa.ParameterSizes]()
	serde.RegisterType[dsa.Parameters]()
	serde.RegisterType[dsa.PrivateKey]()
	serde.RegisterType[dsa.PublicKey]()
	serde.RegisterType[ecdh.PrivateKey]()
	serde.RegisterType[ecdh.PublicKey]()
	serde.RegisterType[ecdsa.PrivateKey]()
	serde.RegisterType[ecdsa.PublicKey]()
	serde.RegisterType[ed25519.Options]()
	serde.RegisterType[ed25519.PrivateKey]()
	serde.RegisterType[ed25519.PublicKey]()
	serde.RegisterType[elliptic.CurveParams]()
	serde.RegisterType[embed.FS]()
	serde.RegisterType[flate.CorruptInputError]()
	serde.RegisterType[flate.InternalError]()
	serde.RegisterType[flate.ReadError]()
	serde.RegisterType[flate.WriteError]()
	serde.RegisterType[flate.Writer]()
	serde.RegisterType[float32]()
	serde.RegisterType[float64]()
	serde.RegisterType[fs.FileMode]()
	serde.RegisterType[fs.PathError]()
	serde.RegisterType[gzip.Header]()
	serde.RegisterType[gzip.Reader]()
	serde.RegisterType[gzip.Writer]()
	serde.RegisterType[hex.InvalidByteError]()
	serde.RegisterType[hpack.Decoder]()
	serde.RegisterType[hpack.DecodingError]()
	serde.RegisterType[hpack.Encoder]()
	serde.RegisterType[hpack.HeaderField]()
	serde.RegisterType[hpack.InvalidIndexError]()
	serde.RegisterType[http.Client]()
	serde.RegisterType[http.ConnState]()
	serde.RegisterType[http.Cookie]()
	serde.RegisterType[http.Dir]()
	serde.RegisterType[http.Header]()
	serde.RegisterType[http.MaxBytesError]()
	serde.RegisterType[http.ProtocolError]()
	serde.RegisterType[http.PushOptions]()
	serde.RegisterType[http.Request]()
	serde.RegisterType[http.Response]()
	serde.RegisterType[http.ResponseController]()
	serde.RegisterType[http.SameSite]()
	serde.RegisterType[http.ServeMux]()
	serde.RegisterType[http.Server]()
	serde.RegisterType[http.Transport]()
	serde.RegisterType[httpproxy.Config]()
	serde.RegisterType[httptrace.ClientTrace]()
	serde.RegisterType[httptrace.DNSDoneInfo]()
	serde.RegisterType[httptrace.DNSStartInfo]()
	serde.RegisterType[httptrace.GotConnInfo]()
	serde.RegisterType[httptrace.WroteRequestInfo]()
	serde.RegisterType[idna.Profile]()
	serde.RegisterType[int]()
	serde.RegisterType[int16]()
	serde.RegisterType[int32]()
	serde.RegisterType[int64]()
	serde.RegisterType[int8]()
	serde.RegisterType[internal.FlushAfterChunkWriter]()
	serde.RegisterType[io.LimitedReader]()
	serde.RegisterType[io.OffsetWriter]()
	serde.RegisterType[io.PipeReader]()
	serde.RegisterType[io.PipeWriter]()
	serde.RegisterType[io.SectionReader]()
	serde.RegisterType[list.Element]()
	serde.RegisterType[list.List]()
	serde.RegisterType[log.Logger]()
	serde.RegisterType[map[*byte][]byte]()
	serde.RegisterType[map[int]int]()
	serde.RegisterType[map[int]string]()
	serde.RegisterType[map[string][]int]()
	serde.RegisterType[map[string][]string]()
	serde.RegisterType[map[string]bool]()
	serde.RegisterType[map[string]int]()
	serde.RegisterType[map[string]map[string]int]()
	serde.RegisterType[map[string]map[string]string]()
	serde.RegisterType[map[string]string]()
	serde.RegisterType[map[string]struct{}]()
	serde.RegisterType[map[string]uint64]()
	serde.RegisterType[map[uint16]bool]()
	serde.RegisterType[map[uint32]rune]()
	serde.RegisterType[map[uint64]bool]()
	serde.RegisterType[mime.WordDecoder]()
	serde.RegisterType[mime.WordEncoder]()
	serde.RegisterType[multipart.FileHeader]()
	serde.RegisterType[multipart.Form]()
	serde.RegisterType[multipart.Part]()
	serde.RegisterType[multipart.Reader]()
	serde.RegisterType[multipart.Writer]()
	serde.RegisterType[net.AddrError]()
	serde.RegisterType[net.Buffers]()
	serde.RegisterType[net.DNSConfigError]()
	serde.RegisterType[net.DNSError]()
	serde.RegisterType[net.Dialer]()
	serde.RegisterType[net.Flags]()
	serde.RegisterType[net.HardwareAddr]()
	serde.RegisterType[net.IP]()
	serde.RegisterType[net.IPAddr]()
	serde.RegisterType[net.IPConn]()
	serde.RegisterType[net.IPMask]()
	serde.RegisterType[net.IPNet]()
	serde.RegisterType[net.Interface]()
	serde.RegisterType[net.InvalidAddrError]()
	serde.RegisterType[net.ListenConfig]()
	serde.RegisterType[net.MX]()
	serde.RegisterType[net.NS]()
	serde.RegisterType[net.OpError]()
	serde.RegisterType[net.ParseError]()
	serde.RegisterType[net.Resolver]()
	serde.RegisterType[net.SRV]()
	serde.RegisterType[net.TCPAddr]()
	serde.RegisterType[net.TCPConn]()
	serde.RegisterType[net.TCPListener]()
	serde.RegisterType[net.UDPAddr]()
	serde.RegisterType[net.UDPConn]()
	serde.RegisterType[net.UnixAddr]()
	serde.RegisterType[net.UnixConn]()
	serde.RegisterType[net.UnixListener]()
	serde.RegisterType[net.UnknownNetworkError]()
	serde.RegisterType[netip.Addr]()
	serde.RegisterType[netip.AddrPort]()
	serde.RegisterType[netip.Prefix]()
	serde.RegisterType[norm.Form]()
	serde.RegisterType[norm.Iter]()
	serde.RegisterType[norm.Properties]()
	serde.RegisterType[os.File]()
	serde.RegisterType[os.LinkError]()
	serde.RegisterType[os.ProcAttr]()
	serde.RegisterType[os.Process]()
	serde.RegisterType[os.ProcessState]()
	serde.RegisterType[os.SyscallError]()
	serde.RegisterType[pem.Block]()
	serde.RegisterType[pkix.AlgorithmIdentifier]()
	serde.RegisterType[pkix.AttributeTypeAndValue]()
	serde.RegisterType[pkix.AttributeTypeAndValueSET]()
	serde.RegisterType[pkix.CertificateList]()
	serde.RegisterType[pkix.Extension]()
	serde.RegisterType[pkix.Name]()
	serde.RegisterType[pkix.RDNSequence]()
	serde.RegisterType[pkix.RelativeDistinguishedNameSET]()
	serde.RegisterType[pkix.RevokedCertificate]()
	serde.RegisterType[pkix.TBSCertificateList]()
	serde.RegisterType[quotedprintable.Reader]()
	serde.RegisterType[quotedprintable.Writer]()
	serde.RegisterType[rand.Rand]()
	serde.RegisterType[rand.Zipf]()
	serde.RegisterType[rc4.Cipher]()
	serde.RegisterType[rc4.KeySizeError]()
	serde.RegisterType[reflect.ChanDir]()
	serde.RegisterType[reflect.Kind]()
	serde.RegisterType[reflect.MapIter]()
	serde.RegisterType[reflect.Method]()
	serde.RegisterType[reflect.SelectCase]()
	serde.RegisterType[reflect.SelectDir]()
	serde.RegisterType[reflect.SliceHeader]()
	serde.RegisterType[reflect.StringHeader]()
	serde.RegisterType[reflect.StructField]()
	serde.RegisterType[reflect.StructTag]()
	serde.RegisterType[reflect.Value]()
	serde.RegisterType[reflect.ValueError]()
	serde.RegisterType[route.DefaultAddr]()
	serde.RegisterType[route.Inet4Addr]()
	serde.RegisterType[route.Inet6Addr]()
	serde.RegisterType[route.InterfaceAddrMessage]()
	serde.RegisterType[route.InterfaceAnnounceMessage]()
	serde.RegisterType[route.InterfaceMessage]()
	serde.RegisterType[route.InterfaceMetrics]()
	serde.RegisterType[route.InterfaceMulticastAddrMessage]()
	serde.RegisterType[route.LinkAddr]()
	serde.RegisterType[route.RIBType]()
	serde.RegisterType[route.RouteMessage]()
	serde.RegisterType[route.RouteMetrics]()
	serde.RegisterType[route.SysType]()
	serde.RegisterType[rsa.CRTValue]()
	serde.RegisterType[rsa.OAEPOptions]()
	serde.RegisterType[rsa.PKCS1v15DecryptOptions]()
	serde.RegisterType[rsa.PSSOptions]()
	serde.RegisterType[rsa.PrecomputedValues]()
	serde.RegisterType[rsa.PrivateKey]()
	serde.RegisterType[rsa.PublicKey]()
	serde.RegisterType[rune]()
	serde.RegisterType[runtime.BlockProfileRecord]()
	serde.RegisterType[runtime.Frame]()
	serde.RegisterType[runtime.Frames]()
	serde.RegisterType[runtime.Func]()
	serde.RegisterType[runtime.MemProfileRecord]()
	serde.RegisterType[runtime.MemStats]()
	serde.RegisterType[runtime.PanicNilError]()
	serde.RegisterType[runtime.Pinner]()
	serde.RegisterType[runtime.StackRecord]()
	serde.RegisterType[runtime.TypeAssertionError]()
	serde.RegisterType[sort.Float64Slice]()
	serde.RegisterType[sort.IntSlice]()
	serde.RegisterType[sort.StringSlice]()
	serde.RegisterType[strconv.NumError]()
	serde.RegisterType[string]()
	serde.RegisterType[strings.Builder]()
	serde.RegisterType[strings.Reader]()
	serde.RegisterType[strings.Replacer]()
	serde.RegisterType[struct{b bool; x any}]()
	serde.RegisterType[struct{base uintptr; end uintptr}]()
	serde.RegisterType[struct{enabled bool; pad [3]byte; needed bool; alignme uint64}]()
	serde.RegisterType[struct{fd int32; cmd int32; arg int32; ret int32; errno int32}]()
	serde.RegisterType[struct{fill uint64; capacity uint64}]()
	serde.RegisterType[struct{fn uintptr; a1 uintptr; a2 uintptr; a3 uintptr; a4 uintptr; a5 uintptr; a6 uintptr; r1 uintptr; r2 uintptr; err uintptr}]()
	serde.RegisterType[struct{fn uintptr; a1 uintptr; a2 uintptr; a3 uintptr; a4 uintptr; a5 uintptr; f1 float64; r1 uintptr}]()
	serde.RegisterType[struct{fn uintptr; a1 uintptr; a2 uintptr; a3 uintptr; r1 uintptr; r2 uintptr; err uintptr}]()
	serde.RegisterType[struct{t int64; numer uint32; denom uint32}]()
	serde.RegisterType[struct{tick uint64; i int}]()
	serde.RegisterType[struct{}]()
	serde.RegisterType[sync.Cond]()
	serde.RegisterType[sync.Map]()
	serde.RegisterType[sync.Mutex]()
	serde.RegisterType[sync.Once]()
	serde.RegisterType[sync.Pool]()
	serde.RegisterType[sync.RWMutex]()
	serde.RegisterType[sync.WaitGroup]()
	serde.RegisterType[syscall.BpfHdr]()
	serde.RegisterType[syscall.BpfInsn]()
	serde.RegisterType[syscall.BpfProgram]()
	serde.RegisterType[syscall.BpfStat]()
	serde.RegisterType[syscall.BpfVersion]()
	serde.RegisterType[syscall.Cmsghdr]()
	serde.RegisterType[syscall.Credential]()
	serde.RegisterType[syscall.Dirent]()
	serde.RegisterType[syscall.Errno]()
	serde.RegisterType[syscall.Fbootstraptransfer_t]()
	serde.RegisterType[syscall.FdSet]()
	serde.RegisterType[syscall.Flock_t]()
	serde.RegisterType[syscall.Fsid]()
	serde.RegisterType[syscall.Fstore_t]()
	serde.RegisterType[syscall.ICMPv6Filter]()
	serde.RegisterType[syscall.IPMreq]()
	serde.RegisterType[syscall.IPv6MTUInfo]()
	serde.RegisterType[syscall.IPv6Mreq]()
	serde.RegisterType[syscall.IfData]()
	serde.RegisterType[syscall.IfMsghdr]()
	serde.RegisterType[syscall.IfaMsghdr]()
	serde.RegisterType[syscall.IfmaMsghdr]()
	serde.RegisterType[syscall.IfmaMsghdr2]()
	serde.RegisterType[syscall.Inet4Pktinfo]()
	serde.RegisterType[syscall.Inet6Pktinfo]()
	serde.RegisterType[syscall.InterfaceAddrMessage]()
	serde.RegisterType[syscall.InterfaceMessage]()
	serde.RegisterType[syscall.InterfaceMulticastAddrMessage]()
	serde.RegisterType[syscall.Iovec]()
	serde.RegisterType[syscall.Kevent_t]()
	serde.RegisterType[syscall.Linger]()
	serde.RegisterType[syscall.Log2phys_t]()
	serde.RegisterType[syscall.Msghdr]()
	serde.RegisterType[syscall.ProcAttr]()
	serde.RegisterType[syscall.Radvisory_t]()
	serde.RegisterType[syscall.RawSockaddr]()
	serde.RegisterType[syscall.RawSockaddrAny]()
	serde.RegisterType[syscall.RawSockaddrDatalink]()
	serde.RegisterType[syscall.RawSockaddrInet4]()
	serde.RegisterType[syscall.RawSockaddrInet6]()
	serde.RegisterType[syscall.RawSockaddrUnix]()
	serde.RegisterType[syscall.Rlimit]()
	serde.RegisterType[syscall.RouteMessage]()
	serde.RegisterType[syscall.RtMetrics]()
	serde.RegisterType[syscall.RtMsghdr]()
	serde.RegisterType[syscall.Rusage]()
	serde.RegisterType[syscall.Signal]()
	serde.RegisterType[syscall.SockaddrDatalink]()
	serde.RegisterType[syscall.SockaddrInet4]()
	serde.RegisterType[syscall.SockaddrInet6]()
	serde.RegisterType[syscall.SockaddrUnix]()
	serde.RegisterType[syscall.SocketControlMessage]()
	serde.RegisterType[syscall.Stat_t]()
	serde.RegisterType[syscall.Statfs_t]()
	serde.RegisterType[syscall.SysProcAttr]()
	serde.RegisterType[syscall.Termios]()
	serde.RegisterType[syscall.Timespec]()
	serde.RegisterType[syscall.Timeval]()
	serde.RegisterType[syscall.Timeval32]()
	serde.RegisterType[syscall.WaitStatus]()
	serde.RegisterType[textproto.Conn]()
	serde.RegisterType[textproto.Error]()
	serde.RegisterType[textproto.MIMEHeader]()
	serde.RegisterType[textproto.Pipeline]()
	serde.RegisterType[textproto.ProtocolError]()
	serde.RegisterType[textproto.Reader]()
	serde.RegisterType[textproto.Writer]()
	serde.RegisterType[time.Duration]()
	serde.RegisterType[time.Location]()
	serde.RegisterType[time.Month]()
	serde.RegisterType[time.ParseError]()
	serde.RegisterType[time.Ticker]()
	serde.RegisterType[time.Time]()
	serde.RegisterType[time.Timer]()
	serde.RegisterType[time.Weekday]()
	serde.RegisterType[tls.AlertError]()
	serde.RegisterType[tls.Certificate]()
	serde.RegisterType[tls.CertificateRequestInfo]()
	serde.RegisterType[tls.CertificateVerificationError]()
	serde.RegisterType[tls.CipherSuite]()
	serde.RegisterType[tls.ClientAuthType]()
	serde.RegisterType[tls.ClientHelloInfo]()
	serde.RegisterType[tls.ClientSessionState]()
	serde.RegisterType[tls.Config]()
	serde.RegisterType[tls.Conn]()
	serde.RegisterType[tls.ConnectionState]()
	serde.RegisterType[tls.CurveID]()
	serde.RegisterType[tls.Dialer]()
	serde.RegisterType[tls.QUICConfig]()
	serde.RegisterType[tls.QUICConn]()
	serde.RegisterType[tls.QUICEncryptionLevel]()
	serde.RegisterType[tls.QUICEvent]()
	serde.RegisterType[tls.QUICEventKind]()
	serde.RegisterType[tls.QUICSessionTicketOptions]()
	serde.RegisterType[tls.RecordHeaderError]()
	serde.RegisterType[tls.RenegotiationSupport]()
	serde.RegisterType[tls.SessionState]()
	serde.RegisterType[tls.SignatureScheme]()
	serde.RegisterType[transform.NopResetter]()
	serde.RegisterType[transform.Reader]()
	serde.RegisterType[transform.Writer]()
	serde.RegisterType[uint]()
	serde.RegisterType[uint16]()
	serde.RegisterType[uint32]()
	serde.RegisterType[uint64]()
	serde.RegisterType[uint8]()
	serde.RegisterType[uintptr]()
	serde.RegisterType[unicode.CaseRange]()
	serde.RegisterType[unicode.Range16]()
	serde.RegisterType[unicode.Range32]()
	serde.RegisterType[unicode.RangeTable]()
	serde.RegisterType[unicode.SpecialCase]()
	serde.RegisterType[unsafe.Pointer]()
	serde.RegisterType[url.Error]()
	serde.RegisterType[url.EscapeError]()
	serde.RegisterType[url.InvalidHostError]()
	serde.RegisterType[url.URL]()
	serde.RegisterType[url.Userinfo]()
	serde.RegisterType[url.Values]()
	serde.RegisterType[x509.CertPool]()
	serde.RegisterType[x509.Certificate]()
	serde.RegisterType[x509.CertificateInvalidError]()
	serde.RegisterType[x509.CertificateRequest]()
	serde.RegisterType[x509.ConstraintViolationError]()
	serde.RegisterType[x509.ExtKeyUsage]()
	serde.RegisterType[x509.HostnameError]()
	serde.RegisterType[x509.InsecureAlgorithmError]()
	serde.RegisterType[x509.InvalidReason]()
	serde.RegisterType[x509.KeyUsage]()
	serde.RegisterType[x509.PEMCipher]()
	serde.RegisterType[x509.PublicKeyAlgorithm]()
	serde.RegisterType[x509.RevocationList]()
	serde.RegisterType[x509.RevocationListEntry]()
	serde.RegisterType[x509.SignatureAlgorithm]()
	serde.RegisterType[x509.SystemRootsError]()
	serde.RegisterType[x509.UnhandledCriticalExtension]()
	serde.RegisterType[x509.UnknownAuthorityError]()
	serde.RegisterType[x509.VerifyOptions]()
}