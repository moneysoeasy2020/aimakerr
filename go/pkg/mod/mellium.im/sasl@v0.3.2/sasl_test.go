// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package sasl_test

import (
	"bytes"
	"crypto/tls"
	"strconv"
	"testing"

	"mellium.im/sasl"
)

// saslStep is from the perspective of a client, challenge is issued by the
// server and resp is the clients response (the first challenge will generally
// be empty because SASL is a client-first protocol).
type saslStep struct {
	challenge []byte
	resp      []byte
	more      bool
	clientErr bool
	serverErr bool
}

type saslTest struct {
	mechanism  sasl.Mechanism
	clientOpts []sasl.Option
	serverOpts []sasl.Option
	perm       func(*sasl.Negotiator) bool
	steps      []saslStep
	skipClient bool
	skipServer bool
}

func getStepName(n negotiator) string {
	switch n.State() & sasl.StepMask {
	case sasl.Initial:
		return "Initial"
	case sasl.AuthTextSent:
		return "AuthTextSent"
	case sasl.ResponseSent:
		return "ResponseSent"
	case sasl.ValidServerResponse:
		return "ValidServerResponse"
	default:
		panic("Step part of state byte apparently has too many bits")
	}
}

var (
	plainResp       = []byte("Ursel\x00Kurt\x00xipj3plmq")
	testNonce       = []byte("fyko+d2lbbFgONRv9qkxdawL")
	plainClientOpts = []sasl.Option{sasl.Credentials(func() ([]byte, []byte, []byte) {
		return []byte("Kurt"), []byte("xipj3plmq"), []byte("Ursel")
	})}
)

func acceptAll(_ *sasl.Negotiator) bool {
	return true
}

var saslTestCases = [...]saslTest{
	0: {
		skipServer: true,
		mechanism:  sasl.Plain,
		clientOpts: plainClientOpts,
		steps: []saslStep{
			{resp: plainResp, more: false},
			{challenge: nil, resp: nil, clientErr: true, more: false},
		},
	},
	1: {
		skipServer: true,
		mechanism:  sasl.ScramSha1,
		clientOpts: []sasl.Option{sasl.Credentials(func() ([]byte, []byte, []byte) {
			return []byte("user"), []byte("pencil"), []byte{}
		})},
		steps: []saslStep{
			{
				resp: []byte(`n,,n=user,r=fyko+d2lbbFgONRv9qkxdawL`),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,s=QSXCR+Q6sek8bf92,i=4096`),
				resp:      []byte(`c=biws,r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,p=v0X8v3Bz2T0CJGbJQyF0X+HI4Ts=`),
				more:      true,
			},
			{
				challenge: []byte(`v=rmF9pqV8S7suAoZWja4dJRkFsKQ=`),
				resp:      nil,
				more:      false,
			},
		},
	},
	2: {
		skipServer: true,
		// sasl.Mechanism is not SCRAM-SHA-1-PLUS, but has connstate and remote mechanisms.
		mechanism: sasl.ScramSha1,
		clientOpts: []sasl.Option{
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte("user"), []byte("pencil"), []byte{}
			}),
			sasl.RemoteMechanisms("SCRAM-SHA-1-PLUS", "SCRAM-SHA-1"),
			sasl.TLSState(tls.ConnectionState{TLSUnique: []byte{0, 1, 2, 3, 4}}),
		},
		steps: []saslStep{
			{
				resp: []byte(`n,,n=user,r=fyko+d2lbbFgONRv9qkxdawL`),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,s=QSXCR+Q6sek8bf92,i=4096`),
				resp:      []byte(`c=biws,r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,p=v0X8v3Bz2T0CJGbJQyF0X+HI4Ts=`),
				more:      true,
			},
			{
				challenge: []byte(`v=rmF9pqV8S7suAoZWja4dJRkFsKQ=`),
				resp:      nil,
				more:      false,
			},
		},
	},
	3: {
		skipServer: true,
		mechanism:  sasl.ScramSha1Plus,
		clientOpts: []sasl.Option{
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte("user"), []byte("pencil"), []byte{}
			}),
			sasl.RemoteMechanisms("SCRAM-SHA-1-PLUS"),
			sasl.TLSState(tls.ConnectionState{Version: tls.VersionTLS11, TLSUnique: []byte{0, 1, 2, 3, 4}}),
		},
		steps: []saslStep{
			{
				resp: []byte(`p=tls-unique,,n=user,r=fyko+d2lbbFgONRv9qkxdawL`),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL16090868851744577,s=QSXCR+Q6sek8bf92,i=4096`),
				resp:      []byte(`c=cD10bHMtdW5pcXVlLCwAAQIDBA==,r=fyko+d2lbbFgONRv9qkxdawL16090868851744577,p=kD6Wfe1kGICYN08YH7oONG2Enb0=`),
				more:      true,
			},
			{
				challenge: []byte(`v=QI0Ihj/QJv+VSyezLtd/d5PrYy0=`),
				resp:      nil,
				more:      false,
			},
		},
	},
	4: {
		skipServer: true,
		mechanism:  sasl.ScramSha256,
		clientOpts: []sasl.Option{sasl.Credentials(func() ([]byte, []byte, []byte) {
			return []byte("user"), []byte("pencil"), []byte{}
		})},
		steps: []saslStep{
			{
				resp: []byte("n,,n=user,r=fyko+d2lbbFgONRv9qkxdawL"),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL%hvYDpWUa2RaTCAfuxFIlj)hNlF$k0,s=W22ZaJ0SNY7soEsUEjb6gQ==,i=4096`),
				resp:      []byte(`c=biws,r=fyko+d2lbbFgONRv9qkxdawL%hvYDpWUa2RaTCAfuxFIlj)hNlF$k0,p=2FUSN0pPcS7P8hBhsxBJOiUDbRoW4KVNGZT0LxVnSek=`),
				more:      true,
			},
			{
				challenge: []byte(`v=zJZjsVp2g+W9jd01vgbsshippfH1sM0tLdBvs+e3DF4=`),
				resp:      nil,
				more:      false,
			},
		},
	},
	5: {
		skipServer: true,
		mechanism:  sasl.ScramSha256Plus,
		clientOpts: []sasl.Option{
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte("user"), []byte("pencil"), []byte("admin")
			}),
			sasl.RemoteMechanisms("SCRAM-SOMETHING", "SCRAM-SHA-256-PLUS"),
			sasl.TLSState(tls.ConnectionState{TLSUnique: []byte{0, 1, 2, 3, 4}}),
		},
		steps: []saslStep{
			{
				resp: []byte("p=tls-unique,a=admin,n=user,r=fyko+d2lbbFgONRv9qkxdawL"),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL,s=W22ZaJ0SNY7soEsUEjb6gQ==,i=4096`),
				resp:      []byte(`c=cD10bHMtdW5pcXVlLGE9YWRtaW4sAAECAwQ=,r=fyko+d2lbbFgONRv9qkxdawL,p=USNVS9hYD1JWfBOQwzc8o/9vFPQ7kA4CKsocmko/8yU=`),
				more:      true,
			},
			{
				challenge: []byte(`v=zjC1aKz20rqp7P92qtiJD1+gihbP5dKzIUFlBWgOuss=`),
				resp:      nil,
				more:      false,
			},
		},
	},
	6: {
		skipServer: true,
		mechanism:  sasl.ScramSha1Plus,
		clientOpts: []sasl.Option{
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte(",=,="), []byte("password"), []byte{}
			}),
			sasl.RemoteMechanisms("SCRAM-SHA-1-PLUS"),
			sasl.TLSState(tls.ConnectionState{TLSUnique: []byte("finishedmessage")}),
		},
		steps: []saslStep{
			{
				resp: []byte("p=tls-unique,,n==2C=3D=2C=3D,r=fyko+d2lbbFgONRv9qkxdawL"),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawLtheirnonce,s=W22ZaJ0SNY7soEsUEjb6gQ==,i=4096`),
				resp:      []byte(`c=cD10bHMtdW5pcXVlLCxmaW5pc2hlZG1lc3NhZ2U=,r=fyko+d2lbbFgONRv9qkxdawLtheirnonce,p=8t6BJnSAd7Vi+mGZEi+Oqwci11c=`),
				more:      true,
			},
			{
				challenge: []byte(`v=8IDvl31piL1lkn6XLCqqFVS4EJM=`),
				resp:      nil,
				more:      false,
			},
		},
	},
	7: {
		skipClient: true,
		mechanism:  sasl.Plain,
		perm:       acceptAll,
		steps: []saslStep{
			{resp: []byte("\x00Ursel\x00Kurt\x00xipj3plmq"), serverErr: true, more: false},
		},
	},
	8: {
		mechanism: sasl.Plain,
		perm: func(n *sasl.Negotiator) bool {
			user, pass, ident := n.Credentials()
			switch {
			case string(user) != "Kurt":
				return false
			case string(pass) != "xipj3plmq":
				return false
			case string(ident) != "Ursel":
				return false
			}
			return true
		},
		clientOpts: plainClientOpts,
		// serverOpts is only set to smuggle the password into integration tests.
		serverOpts: plainClientOpts,
		steps: []saslStep{
			{resp: plainResp, more: false},
		},
	},
	9: {
		mechanism: sasl.Plain,
		perm: func(n *sasl.Negotiator) bool {
			user, _, _ := n.Credentials()
			return string(user) == "FAIL"
		},
		clientOpts: plainClientOpts,
		steps: []saslStep{
			{resp: plainResp, serverErr: true, more: false},
		},
	},
	10: {
		mechanism: sasl.Plain,
		perm: func(n *sasl.Negotiator) bool {
			_, pass, _ := n.Credentials()
			return string(pass) == "FAIL"
		},
		clientOpts: plainClientOpts,
		steps: []saslStep{
			{resp: plainResp, serverErr: true, more: false},
		},
	},
	11: {
		mechanism: sasl.Plain,
		perm: func(n *sasl.Negotiator) bool {
			_, _, ident := n.Credentials()
			return string(ident) == "FAIL"
		},
		clientOpts: plainClientOpts,
		steps: []saslStep{
			{resp: plainResp, serverErr: true, more: false},
		},
	},
	12: {
		mechanism:  sasl.Plain,
		clientOpts: plainClientOpts,
		steps: []saslStep{
			{resp: plainResp, serverErr: true, more: false},
		},
	},
	13: {
		skipClient: true,
		mechanism:  sasl.Plain,
		perm:       acceptAll,
		steps: []saslStep{
			{resp: []byte("Ursel\x00Kurt\x00xipj3plmq\x00"), serverErr: true, more: false},
		},
	},
	14: {
		skipServer: true,
		mechanism:  sasl.ScramSha256Plus,
		clientOpts: []sasl.Option{
			sasl.RemoteMechanisms("SCRAM-SHA-256-PLUS"),
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte("user"), []byte("pencil"), []byte("admin")
			}),
			sasl.TLSState(tls.ConnectionState{Version: tls.VersionTLS12, TLSUnique: []byte{}}),
		},
		steps: []saslStep{
			{
				resp: []byte("p=tls-unique,a=admin,n=user,r=fyko+d2lbbFgONRv9qkxdawL"),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL%hvYDpWUa2RaTCAfuxFIlj)hNlF$k0,s=W22ZaJ0SNY7soEsUEjb6gQ==,i=4096`),
				clientErr: true,
			},
		},
	},
	15: {
		skipServer: true,
		mechanism:  sasl.ScramSha256Plus,
		clientOpts: []sasl.Option{
			sasl.RemoteMechanisms("SCRAM-SHA-256-PLUS"),
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte("user"), []byte("pencil"), []byte("admin")
			}),
			sasl.TLSState(tls.ConnectionState{TLSUnique: nil}),
		},
		steps: []saslStep{
			{
				resp: []byte("p=tls-unique,a=admin,n=user,r=fyko+d2lbbFgONRv9qkxdawL"),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL%hvYDpWUa2RaTCAfuxFIlj)hNlF$k0,s=W22ZaJ0SNY7soEsUEjb6gQ==,i=4096`),
				clientErr: true,
			},
		},
	},
	16: {
		skipServer: true,
		mechanism:  sasl.ScramSha256Plus,
		clientOpts: []sasl.Option{
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte("user"), []byte("pencil"), []byte("admin")
			}),
			sasl.RemoteMechanisms("SCRAM-SHA-256-PLUS"),
		},
		steps: []saslStep{
			{
				resp: []byte("n,a=admin,n=user,r=fyko+d2lbbFgONRv9qkxdawL"),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL%hvYDpWUa2RaTCAfuxFIlj)hNlF$k0,s=W22ZaJ0SNY7soEsUEjb6gQ==,i=4096`),
				clientErr: true,
			},
		},
	},
	17: {
		skipServer: true,
		// sasl.Mechanism is not SCRAM-SHA-1-PLUS, but has TLS 1.3 connstate and remote
		// mechanisms.
		mechanism: sasl.ScramSha1,
		clientOpts: []sasl.Option{
			sasl.Credentials(func() ([]byte, []byte, []byte) {
				return []byte("user"), []byte("pencil"), []byte{}
			}),
			sasl.RemoteMechanisms("SCRAM-SHA-1-PLUS", "SCRAM-SHA-1"),
			sasl.TLSState(tls.ConnectionState{Version: tls.VersionTLS13}),
		},
		steps: []saslStep{
			{
				resp: []byte(`n,,n=user,r=fyko+d2lbbFgONRv9qkxdawL`),
				more: true,
			},
			{
				challenge: []byte(`r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,s=QSXCR+Q6sek8bf92,i=4096`),
				resp:      []byte(`c=biws,r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,p=v0X8v3Bz2T0CJGbJQyF0X+HI4Ts=`),
				more:      true,
			},
			{
				challenge: []byte(`v=rmF9pqV8S7suAoZWja4dJRkFsKQ=`),
				resp:      nil,
				more:      false,
			},
		},
	},
}

type negotiator interface {
	Step(challenge []byte) (more bool, resp []byte, err error)
	State() sasl.State
}

func testClient(t *testing.T, client *sasl.Negotiator, tc saslTest, run int) {
	for _, step := range tc.steps {
		more, resp, err := client.Step(step.challenge)
		t.Logf("Run %d, Step %s", run, getStepName(client))
		switch {
		case err != nil && client.State()&sasl.Errored != sasl.Errored:
			t.Fatalf("State machine internal error state was not set, got error: %v", err)
		case err == nil && client.State()&sasl.Errored == sasl.Errored:
			t.Fatal("State machine internal error state was set, but no error was returned")
		case err == nil && step.clientErr:
			// There was no error, but we expect one
			t.Fatal("Expected SASL step to error")
		case err != nil && !step.clientErr:
			// There was an error, but we didn't expect one
			t.Fatalf("Got unexpected SASL error: %v", err)
		case !bytes.Equal(step.resp, resp):
			t.Fatalf("Got invalid response text:\nexpected `%s'\n     got `%s'", step.resp, resp)
		case more != step.more:
			t.Fatalf("Got unexpected value for more: %v", more)
		}
	}
}

func testServer(t *testing.T, server *sasl.Negotiator, tc saslTest, run int) {
	for _, step := range tc.steps {
		more, challenge, err := server.Step(step.resp)
		t.Logf("Run %d, Step %s", run, getStepName(server))
		switch {
		case err != nil && server.State()&sasl.Errored != sasl.Errored:
			t.Fatalf("State machine internal error state was not set, got error: %v", err)
		case err == nil && server.State()&sasl.Errored == sasl.Errored:
			t.Fatal("State machine internal error state was set, but no error was returned")
		case err == nil && step.serverErr:
			// There was no error, but we expect one
			t.Fatal("Expected SASL step to error")
		case err != nil && !step.serverErr:
			// There was an error, but we didn't expect one
			t.Fatalf("Got unexpected SASL error: %v", err)
		case string(step.challenge) != string(challenge):
			t.Fatalf("Got invalid challenge text:\nexpected `%s'\n     got `%s'", step.challenge, challenge)
		case more != step.more:
			t.Fatalf("Got unexpected value for more: %v", more)
		}
	}
}

func TestSASL(t *testing.T) {
	for i, tc := range saslTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			client := sasl.NewClient(tc.mechanism, tc.clientOpts...)
			if len(client.Nonce()) == 0 {
				t.Fatal("test client did not set nonce!")
			}
			server := sasl.NewServer(tc.mechanism, tc.perm, tc.serverOpts...)
			if len(client.Nonce()) == 0 {
				t.Fatal("test server did not set nonce!")
			}

			// Run each test twice to make sure that Reset actually sets the state
			// back to the initial state.
			for run := 1; run < 3; run++ {
				sasl.Nonce(testNonce)(client)
				sasl.Nonce(testNonce)(server)
				t.Run("Client", func(t *testing.T) {
					if tc.skipClient {
						t.Skip("no client side defined for test")
					}
					testClient(t, client, tc, run)
				})
				t.Run("Server", func(t *testing.T) {
					if tc.skipServer {
						t.Skip("no server side defined for test")
					}
					testServer(t, server, tc, run)
				})

				client.Reset()
				server.Reset()
			}
		})
	}
}
