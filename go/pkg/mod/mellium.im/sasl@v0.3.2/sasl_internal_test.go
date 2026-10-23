// Copyright 2022 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package sasl

// Nonce is an exported version of the setNonce option that is only available in
// tests.
// This is a work around for the fact that we need to be able to have less
// randomness in tests if we want to check the exact output of each step, but we
// don't want to expose this option to the users or test internal implementation
// details (generally speaking).
// Instead we make this an edge-case.
var Nonce = setNonce
