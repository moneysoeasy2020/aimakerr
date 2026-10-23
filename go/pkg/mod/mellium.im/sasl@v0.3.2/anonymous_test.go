// Copyright 2024 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package sasl_test

import (
	"testing"

	"mellium.im/sasl"
)

func TestAnonymous(t *testing.T) {
	// A server that accepts SASL ANONYMOUS unconditionally.
	s := sasl.NewServer(sasl.Anonymous, func(*sasl.Negotiator) bool {
		return true
	})

	c := sasl.NewClient(sasl.Anonymous)

	_, resp, err := c.Step(nil)
	if err != nil {
		t.Fatalf("Client failed: %s", err)
	}

	_, _, err = s.Step(resp)
	if err != nil {
		t.Fatalf("Server failed: %s", err)
	}

	_, _, err = c.Step(nil)
	if err == nil {
		t.Fatalf("Client did not return ErrTooManySteps")
	}

	_, _, err = s.Step(nil)
	if err == nil {
		t.Fatalf("Server did not return ErrTooManySteps")
	}
}
