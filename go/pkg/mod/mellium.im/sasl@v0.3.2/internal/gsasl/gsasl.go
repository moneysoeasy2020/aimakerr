// Copyright 2022 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// Package gsasl uses cgo to facilitate integration testing against libgsasl.
package gsasl // import "mellium.im/sasl/internal/gsasl"

/*
#cgo LDFLAGS: -lgsasl
#include <gsasl.h>
#include <stdlib.h>

static int cb(Gsasl *ctx, Gsasl_session *sctx, Gsasl_property prop) {
	int rc = GSASL_NO_CALLBACK;

	switch (prop) {
	case GSASL_PASSWORD:
		void *vpass = gsasl_callback_hook_get(ctx);
		gsasl_property_set (sctx, GSASL_PASSWORD, (char*)vpass);
		break;
	default:
		break;
	}

	return rc;
}

void _register_cb(Gsasl *ctx) {
	gsasl_callback_set(ctx, cb);
}
*/
import "C"
import (
	"encoding/base64"
	"fmt"
	"runtime"
	"unsafe"

	"mellium.im/sasl"
)

func saslErr(rc C.int) error {
	msg := C.GoString(C.gsasl_strerror(rc))
	return fmt.Errorf("error %d: %s", rc, msg)
}

// Session represents a SASL client or server.
type Session struct {
	session *C.Gsasl_session
	state   sasl.State
	pass    unsafe.Pointer
}

// NewClient creates a new gsasl client using the options from a Mellium SASL
// client.
func NewClient(m sasl.Mechanism, opts ...sasl.Option) (*Session, error) {
	// Creating a sasl.Negotiator just to apply options and discard it is a bit
	// jank, but it lets us keep the API nice while using it for something it
	// really wasn't designed for.
	n := sasl.NewClient(m, opts...)
	saslCtx := newGSASL()
	var s *C.Gsasl_session
	runtime.SetFinalizer(s, func(s *C.Gsasl_session) {
		C.gsasl_finish(s)
	})

	cName := C.CString(m.Name)
	defer C.free(unsafe.Pointer(cName))
	if rc := C.gsasl_client_start(saslCtx, cName, &s); rc != C.GSASL_OK {
		return nil, fmt.Errorf("gsasl: failed to initialize client: %w", saslErr(rc))
	}

	setProps(n, s)

	c := &Session{
		session: s,
	}
	return c, nil
}

// Step advances the client and returns a result in response to a SASL
// challenge.
func (c *Session) Step(challenge []byte) (more bool, resp []byte, err error) {
	if c.state&sasl.Errored == sasl.Errored {
		panic("gsasl: Step called on a SASL state machine that has errored")
	}
	defer func() {
		if err != nil {
			c.state |= sasl.Errored
		}
	}()

	var cBuf *C.char
	b64In := base64.StdEncoding.EncodeToString(challenge)
	cIn := C.CString(b64In)
	rc := C.gsasl_step64(c.session, cIn, &cBuf)
	if rc != C.GSASL_NEEDS_MORE && rc != C.GSASL_OK {
		return false, nil, fmt.Errorf("gsasl: failed negotiation step: %w", saslErr(rc))
	}
	b64Resp := C.GoString(cBuf)
	C.free(unsafe.Pointer(cIn))
	C.gsasl_free(unsafe.Pointer(cBuf))
	resp, err = base64.StdEncoding.DecodeString(b64Resp)
	if err != nil {
		return false, nil, err
	}

	switch c.state & sasl.StepMask {
	case sasl.Initial:
		c.state = c.state&^sasl.StepMask | sasl.AuthTextSent
	case sasl.AuthTextSent:
		c.state = c.state&^sasl.StepMask | sasl.ResponseSent
	case sasl.ResponseSent:
		c.state = c.state&^sasl.StepMask | sasl.ValidServerResponse
	case sasl.ValidServerResponse:
		return false, nil, sasl.ErrTooManySteps
	}

	if rc != C.GSASL_NEEDS_MORE {
		c.state = c.state&^sasl.StepMask | sasl.ValidServerResponse
	}

	return rc == C.GSASL_NEEDS_MORE, resp, err
}

// Close cleans up after the session.
func (c *Session) Close() error {
	C.gsasl_finish(c.session)
	C.free(c.pass)
	return nil
}

// State returns the internal state in a way that is compatible with
// sasl.Negotiators.
func (c *Session) State() sasl.State {
	return c.state
}

// NewServer creates a new gsasl server using the options from a Mellium SASL
// server.
func NewServer(m sasl.Mechanism, permissions func(*sasl.Negotiator) bool, opts ...sasl.Option) (*Session, error) {
	n := sasl.NewServer(m, permissions, opts...)

	saslCtx := newGSASL()
	var s *C.Gsasl_session

	cName := C.CString(m.Name)
	defer C.free(unsafe.Pointer(cName))

	if rc := C.gsasl_server_start(saslCtx, cName, &s); rc != C.GSASL_OK {
		return nil, fmt.Errorf("failed to initialize gsasl server: %w", saslErr(rc))
	}

	_, pass, _ := n.Credentials()

	c := &Session{
		session: s,
		pass:    unsafe.Pointer(C.CString(string(pass))),
	}
	C.gsasl_callback_hook_set(saslCtx, unsafe.Pointer(c.pass))
	C._register_cb(saslCtx)
	return c, nil
}

func setProps(n *sasl.Negotiator, s *C.Gsasl_session) {
	user, pass, ident := n.Credentials()
	cUser, cPass, cIdent := C.CString(string(user)), C.CString(string(pass)), C.CString(string(ident))
	C.gsasl_property_set(s, C.GSASL_AUTHID, cUser)
	C.gsasl_property_set(s, C.GSASL_PASSWORD, cPass)
	C.gsasl_property_set(s, C.GSASL_AUTHZID, cIdent)
	C.free(unsafe.Pointer(cUser))
	C.free(unsafe.Pointer(cPass))
	C.free(unsafe.Pointer(cIdent))

	if cs := n.TLSState(); cs != nil {
		//lint:ignore SA1019 TLS unique must be supported by SCRAM
		b64Unique := base64.StdEncoding.EncodeToString(cs.TLSUnique)
		cUnique := C.CString(b64Unique)
		C.gsasl_property_set(s, C.GSASL_CB_TLS_UNIQUE, cUnique)
		C.free(unsafe.Pointer(cUnique))
	}
	if nonce := n.Nonce(); len(nonce) > 0 {
		cNonce := C.CString(string(nonce))
		C.gsasl_property_set(s, C.GSASL_SCRAM_SALT, cNonce)
		C.free(unsafe.Pointer(cNonce))
	}
}

func newGSASL() *C.Gsasl {
	var saslCtx *C.Gsasl
	rc := C.gsasl_init(&saslCtx)
	if rc != C.GSASL_OK {
		panic(fmt.Errorf("SASL initialization failed: %w", saslErr(rc)))
	}
	return saslCtx
}
