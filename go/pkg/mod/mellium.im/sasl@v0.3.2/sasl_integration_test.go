// Copyright 2022 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

//go:build integration
// +build integration

package sasl_test

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"mellium.im/sasl"
	"mellium.im/sasl/internal/gsasl"
)

func TestIntegrationGSASL(t *testing.T) {
	for i, tc := range saslTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Run("Client", func(t *testing.T) {
				switch {
				case tc.skipClient:
					t.Skip("no client side defined for integration test")
				case strings.HasPrefix(tc.mechanism.Name, "SCRAM-"):
					t.Skip("server side of SCRAM not implemented")
				}
				client, err := gsasl.NewClient(tc.mechanism, tc.clientOpts...)
				if err != nil {
					t.Fatalf("error creating gsasl client: %v", err)
				}
				/* #nosec */
				defer client.Close()
				testClientIntegration(t, client, tc)
			})
			t.Run("Server", func(t *testing.T) {
				if tc.skipServer {
					t.Skip("no server side defined for integration test")
				}
				server, err := gsasl.NewServer(tc.mechanism, tc.perm, tc.serverOpts...)
				if err != nil {
					t.Fatalf("error creating gsasl server: %v", err)
				}
				/* #nosec */
				defer server.Close()
				testServerIntegration(t, server, tc)
			})
		})
	}
}

func testClientIntegration(t *testing.T, client negotiator, tc saslTest) {
	melClient := sasl.NewClient(tc.mechanism, tc.clientOpts...)

	for _, step := range tc.steps {
		more, resp, err := client.Step(step.challenge)
		melMore, melResp, melErr := melClient.Step(step.challenge)
		t.Logf("step %s", getStepName(client))
		switch {
		case (melErr == nil && err != nil) || (err == nil && melErr != nil):
			t.Errorf("mellium and external library differ in error output: us=%v, them=%v", melErr, err)
		case !bytes.Equal(melResp, resp):
			t.Fatalf("mellium and external library differ in response:\nus   `%s'\nthem `%s'", melResp, resp)
		case err != nil && client.State()&sasl.Errored != sasl.Errored:
			t.Fatalf("state machine internal error state was not set, got error: %v", err)
		case err == nil && client.State()&sasl.Errored == sasl.Errored:
			t.Fatal("state machine internal error state was set, but no error was returned")
		case err == nil && step.clientErr:
			// There was no error, but we expect one
			t.Fatal("expected SASL step to error")
		case err != nil && !step.clientErr:
			// There was an error, but we didn't expect one
			t.Fatalf("got unexpected SASL error: %v", err)
		case more != step.more || melMore != step.more:
			t.Fatalf("got unexpected value for more: want=%v, us=%t, them=%t", step.more, melMore, more)
		}
	}
}

func testServerIntegration(t *testing.T, server negotiator, tc saslTest) {
	melServer := sasl.NewServer(tc.mechanism, tc.perm, tc.serverOpts...)

	for _, step := range tc.steps {
		more, challenge, err := server.Step(step.resp)
		melMore, melChallenge, melErr := melServer.Step(step.resp)
		t.Logf("step %s", getStepName(server))
		switch {
		case (melErr == nil && err != nil) || (err == nil && melErr != nil):
			t.Errorf("mellium and external library differ in error output: us=%v, them=%v", melErr, err)
		case !bytes.Equal(challenge, melChallenge):
			t.Fatalf("mellium and external library differ in issued challenge:\nus   `%s'\nthem `%s'", melChallenge, challenge)
		case err != nil && server.State()&sasl.Errored != sasl.Errored:
			t.Fatalf("state machine internal error state was not set, got error: %v", err)
		case err == nil && server.State()&sasl.Errored == sasl.Errored:
			t.Fatal("state machine internal error state was set, but no error was returned")
		case err == nil && step.serverErr:
			// There was no error, but we expect one
			t.Fatal("expected SASL step to error")
		case err != nil && !step.serverErr:
			// There was an error, but we didn't expect one
			t.Fatalf("got unexpected SASL error: %v", err)
		case more != step.more || melMore != step.more:
			t.Fatalf("got unexpected value for more: want=%t, us=%t, them=%t", step.more, melMore, more)
		}
	}
}
