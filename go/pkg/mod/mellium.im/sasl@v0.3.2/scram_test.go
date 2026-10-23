// Copyright 2023 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package sasl

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// todo: test what will happen with server without SaltedCredentials()
func TestScram(t *testing.T) {
	// This is not very good test since we test both client and server at the same time
	// but it is better than nothing.
	t.Run("sha_256", func(t *testing.T) {
		salt := []byte("salt")
		pass := []byte("123")
		iter := 123
		saltedPassword := SCRAMSaltPassword(sha256.New, pass, salt, iter)
		clientCreds := func() (u []byte, p []byte, i []byte) {
			return []byte("mario"), pass, []byte("")
		}
		saltedCredentials := func(Username, Identity []byte, mechanismName string) ([]byte, []byte, int64, error) {
			return salt, saltedPassword, int64(iter), nil
		}
		clientNegotiator := NewClient(ScramSha256, setNonce([]byte("abcdefghijklmnop")), Credentials(clientCreds))
		serverNegotiator := NewServer(ScramSha256, nil, setNonce([]byte("0123456789012345")), SaltedCredentials(saltedCredentials))

		clientMore, firstClientResp, err := clientNegotiator.Step(nil)
		if err != nil {
			t.Fatalf("client negotiator finished with error on first step: %s", err)
		}
		if !clientMore {
			t.Errorf("client negotiator after first step expected to finish with more=true, but actually finished with more=false")
		}
		expectedFirstClientResp := []byte("n,,n=mario,r=abcdefghijklmnop")
		if !bytes.Equal(expectedFirstClientResp, firstClientResp) {
			t.Errorf("client negotiator after first step expected to finish with response=%s, but actually finished with response=%s", expectedFirstClientResp, firstClientResp)
		}

		serverMore, firstServerResp, err := serverNegotiator.Step(firstClientResp)
		if err != nil {
			t.Fatalf("server negotiator finished with error: %s", err)
		}
		if !serverMore {
			t.Errorf("server negotiator after first step expected to finish with more=true, but actually finished with more=false")
		}
		expectedFirstServerResp := []byte("r=abcdefghijklmnop0123456789012345,s=c2FsdA==,i=123")
		if !bytes.Equal(expectedFirstServerResp, firstServerResp) {
			t.Errorf("server negotiator after first step expected to finish with response=%s, but actually finished with response=%s", expectedFirstServerResp, firstServerResp)
		}

		clientMore, finalClientResp, err := clientNegotiator.Step(firstServerResp)
		if err != nil {
			t.Fatalf("client negotiator finished with error on final step: %s", err)
		}
		if !clientMore {
			t.Errorf("client negotiator after final step expected to finish with more=true, but actually finished with more=false")
		}
		expectedFinalClientResp := []byte("c=biws,r=abcdefghijklmnop0123456789012345,p=qT5Jb07VyUR1i/BSy/IzaG54XvkHiO9fiSqwJBvkYxE=")
		if !bytes.Equal(expectedFinalClientResp, finalClientResp) {
			t.Errorf("client negotiator after final step expected to finish with response=%s, but actually finished with response=%s", expectedFinalClientResp, finalClientResp)
		}

		serverMore, finalServerResp, err := serverNegotiator.Step(finalClientResp)
		if err != nil {
			t.Fatalf("server negotiator finished with error on final step: %s", err)
		}
		if serverMore {
			t.Errorf("server negotiator after final step expected to finish with more=false, but actually finished with more=true")
		}
		expectedFinalServerResp := []byte("v=yfna63FbW3txQzHcnNVdZsavvFTo0FZzA0ymVYk/Tkk=")
		if !bytes.Equal(expectedFinalServerResp, finalServerResp) {
			t.Errorf("server negotiator after final step expected to finish with response=%s, but actually finished with response=%s", expectedFinalServerResp, finalServerResp)
		}

		clientMore, emptyClientResp, err := clientNegotiator.Step(finalServerResp)
		if err != nil {
			t.Fatalf("client negotiator finished with error on terminating step: %s", err)
		}
		if clientMore {
			t.Errorf("client negotiator after terminating step expected to finish with more=true, but actually finished with more=false")
		}
		if len(emptyClientResp) != 0 {
			t.Errorf("client negotiator after terminating step expected to finish with empty response, but actually finished with response=%s", emptyClientResp)
		}
	})

	t.Run("sha_256_invalid_client_name", func(t *testing.T) {
		salt := []byte("salt")
		pass := []byte("123")
		iter := 123
		saltedPassword := SCRAMSaltPassword(sha256.New, pass, salt, iter)
		clientCreds := func() (u []byte, p []byte, i []byte) {
			return []byte("mario="), pass, []byte("")
		}
		saltedCredentials := func(Username, Identity []byte, mechanismName string) ([]byte, []byte, int64, error) {
			return salt, saltedPassword, int64(iter), nil
		}
		clientNegotiator := NewClient(ScramSha256, setNonce([]byte("abcdefghijklmnop")), Credentials(clientCreds))
		serverNegotiator := NewServer(ScramSha256, nil, setNonce([]byte("0123456789012345")), SaltedCredentials(saltedCredentials))

		clientMore, firstClientResp, err := clientNegotiator.Step(nil)
		if err != nil {
			t.Fatalf("client negotiator finished with error on first step: %s", err)
		}
		if !clientMore {
			t.Errorf("client negotiator after first step expected to finish with more=true, but actually finished with more=false")
		}
		expectedFirstClientResp := []byte("n,,n=mario=3D,r=abcdefghijklmnop")
		if !bytes.Equal(expectedFirstClientResp, firstClientResp) {
			t.Errorf("client negotiator after first step expected to finish with response=%s, but actually finished with response=%s", expectedFirstClientResp, firstClientResp)
		}

		serverMore, firstServerResp, err := serverNegotiator.Step(firstClientResp)
		expectedErr := fmt.Errorf("unescaped username contains '='")
		if err == nil {
			t.Errorf("server negotiator after first step should return %q error, but actually returned nil error", expectedErr)
		}
		if serverMore {
			t.Errorf("server negotiator after first step expected to finish with more=false, but actually finished with more=true")
		}
		if len(firstServerResp) > 0 {
			t.Errorf("server negotiator after first step expected to finish with empty response, but actually finished with response=%s", firstServerResp)
		}

	})
}

// parseClientFirstMessage test cases

func TestParseClientFirstMessage(t *testing.T) {
	t.Run("positive_flags", func(t *testing.T) {
		message, err := parseClientFirstMessage([]byte("y,,n=mario,r=nonce"))
		if err != nil {
			t.Fatalf("parseClientFirstMessage finished with unexpected error: %s", err)
		}
		expectedGs2CbindFlag := []byte("y")
		expectedAuthzID := []byte(nil)
		expectedUsername := []byte("mario")
		expectedNonce := []byte("nonce")
		expectedGs2Header := []byte("y,,")
		expectedBare := []byte("n=mario,r=nonce")
		if !bytes.Equal(expectedGs2CbindFlag, message.gs2CbindFlag) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2CbindFlag=%s, but actually finished with gs2CbindFlag=%s", expectedGs2CbindFlag, message.gs2CbindFlag)
		}
		if !bytes.Equal(expectedAuthzID, message.authzID) {
			t.Errorf("parseClientFirstMessage expected to finish with authzID=%s, but actually finished with authzID=%s", expectedAuthzID, message.authzID)
		}
		if !bytes.Equal(expectedUsername, message.username) {
			t.Errorf("parseClientFirstMessage expected to finish with username=%s, but actually finished with username=%s", expectedUsername, message.username)
		}
		if !bytes.Equal(expectedNonce, message.nonce) {
			t.Errorf("parseClientFirstMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
		}
		if !bytes.Equal(expectedGs2Header, message.gs2Header) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2Header=%s, but actually finished with gs2Header=%s", expectedGs2Header, message.gs2Header)
		}
		if !bytes.Equal(expectedBare, message.bare) {
			t.Errorf("parseClientFirstMessage expected to finish with bare=%s, but actually finished with bare=%s", expectedBare, message.bare)
		}

		message, err = parseClientFirstMessage([]byte("n,,n=wario,r=alsononce"))
		if err != nil {
			t.Fatalf("parseClientFirstMessage finished with unexpected error: %s", err)
		}
		expectedGs2CbindFlag = []byte("n")
		expectedAuthzID = []byte(nil)
		expectedUsername = []byte("wario")
		expectedNonce = []byte("alsononce")
		expectedGs2Header = []byte("n,,")
		expectedBare = []byte("n=wario,r=alsononce")
		if !bytes.Equal(expectedGs2CbindFlag, message.gs2CbindFlag) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2CbindFlag=%s, but actually finished with gs2CbindFlag=%s", expectedGs2CbindFlag, message.gs2CbindFlag)
		}
		if !bytes.Equal(expectedAuthzID, message.authzID) {
			t.Errorf("parseClientFirstMessage expected to finish with authzID=%s, but actually finished with authzID=%s", expectedAuthzID, message.authzID)
		}
		if !bytes.Equal(expectedUsername, message.username) {
			t.Errorf("parseClientFirstMessage expected to finish with username=%s, but actually finished with username=%s", expectedUsername, message.username)
		}
		if !bytes.Equal(expectedNonce, message.nonce) {
			t.Errorf("parseClientFirstMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
		}
		if !bytes.Equal(expectedGs2Header, message.gs2Header) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2Header=%s, but actually finished with gs2Header=%s", expectedGs2Header, message.gs2Header)
		}
		if !bytes.Equal(expectedBare, message.bare) {
			t.Errorf("parseClientFirstMessage expected to finish with bare=%s, but actually finished with bare=%s", expectedBare, message.bare)
		}

		message, err = parseClientFirstMessage([]byte("p=cb1,,n=wario,r=alsononce"))
		if err != nil {
			t.Fatalf("parseClientFirstMessage finished with unexpected error: %s", err)
		}
		expectedGs2CbindFlag = []byte("p=cb1")
		expectedAuthzID = []byte(nil)
		expectedUsername = []byte("wario")
		expectedNonce = []byte("alsononce")
		expectedGs2Header = []byte("p=cb1,,")
		expectedBare = []byte("n=wario,r=alsononce")
		if !bytes.Equal(expectedGs2CbindFlag, message.gs2CbindFlag) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2CbindFlag=%s, but actually finished with gs2CbindFlag=%s", expectedGs2CbindFlag, message.gs2CbindFlag)
		}
		if !bytes.Equal(expectedAuthzID, message.authzID) {
			t.Errorf("parseClientFirstMessage expected to finish with authzID=%s, but actually finished with authzID=%s", expectedAuthzID, message.authzID)
		}
		if !bytes.Equal(expectedUsername, message.username) {
			t.Errorf("parseClientFirstMessage expected to finish with username=%s, but actually finished with username=%s", expectedUsername, message.username)
		}
		if !bytes.Equal(expectedNonce, message.nonce) {
			t.Errorf("parseClientFirstMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
		}
		if !bytes.Equal(expectedGs2Header, message.gs2Header) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2Header=%s, but actually finished with gs2Header=%s", expectedGs2Header, message.gs2Header)
		}
		if !bytes.Equal(expectedBare, message.bare) {
			t.Errorf("parseClientFirstMessage expected to finish with bare=%s, but actually finished with bare=%s", expectedBare, message.bare)
		}
	})

	t.Run("positive_authzID", func(t *testing.T) {
		message, err := parseClientFirstMessage([]byte("n,a=wario,n=mario,r=nonce"))
		if err != nil {
			t.Fatalf("parseClientFirstMessage finished with unexpected error: %s", err)
		}
		expectedGs2CbindFlag := []byte("n")
		expectedAuthzID := []byte("wario")
		expectedUsername := []byte("mario")
		expectedNonce := []byte("nonce")
		expectedGs2Header := []byte("n,a=wario,")
		expectedBare := []byte("n=mario,r=nonce")
		if !bytes.Equal(expectedGs2CbindFlag, message.gs2CbindFlag) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2CbindFlag=%s, but actually finished with gs2CbindFlag=%s", expectedGs2CbindFlag, message.gs2CbindFlag)
		}
		if !bytes.Equal(expectedAuthzID, message.authzID) {
			t.Errorf("parseClientFirstMessage expected to finish with authzID=%s, but actually finished with authzID=%s", expectedAuthzID, message.authzID)
		}
		if !bytes.Equal(expectedUsername, message.username) {
			t.Errorf("parseClientFirstMessage expected to finish with username=%s, but actually finished with username=%s", expectedUsername, message.username)
		}
		if !bytes.Equal(expectedNonce, message.nonce) {
			t.Errorf("parseClientFirstMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
		}
		if !bytes.Equal(expectedGs2Header, message.gs2Header) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2Header=%s, but actually finished with gs2Header=%s", expectedGs2Header, message.gs2Header)
		}
		if !bytes.Equal(expectedBare, message.bare) {
			t.Errorf("parseClientFirstMessage expected to finish with bare=%s, but actually finished with bare=%s", expectedBare, message.bare)
		}
	})

	t.Run("positive_username", func(t *testing.T) {
		message, err := parseClientFirstMessage([]byte("n,a=😂,n=mario=3D,r=nonce"))
		if err != nil {
			t.Fatalf("parseClientFirstMessage finished with unexpected error: %s", err)
		}
		expectedGs2CbindFlag := []byte("n")
		expectedAuthzID := []byte("😂")
		expectedUsername := []byte("mario=")
		expectedNonce := []byte("nonce")
		expectedGs2Header := []byte("n,a=😂,")
		expectedBare := []byte("n=mario=3D,r=nonce")
		if !bytes.Equal(expectedGs2CbindFlag, message.gs2CbindFlag) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2CbindFlag=%s, but actually finished with gs2CbindFlag=%s", expectedGs2CbindFlag, message.gs2CbindFlag)
		}
		if !bytes.Equal(expectedAuthzID, message.authzID) {
			t.Errorf("parseClientFirstMessage expected to finish with authzID=%s, but actually finished with authzID=%s", expectedAuthzID, message.authzID)
		}
		if !bytes.Equal(expectedUsername, message.username) {
			t.Errorf("parseClientFirstMessage expected to finish with username=%s, but actually finished with username=%s", expectedUsername, message.username)
		}
		if !bytes.Equal(expectedNonce, message.nonce) {
			t.Errorf("parseClientFirstMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
		}
		if !bytes.Equal(expectedGs2Header, message.gs2Header) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2Header=%s, but actually finished with gs2Header=%s", expectedGs2Header, message.gs2Header)
		}
		if !bytes.Equal(expectedBare, message.bare) {
			t.Errorf("parseClientFirstMessage expected to finish with bare=%s, but actually finished with bare=%s", expectedBare, message.bare)
		}
	})

	t.Run("positive_extensions", func(t *testing.T) {
		message, err := parseClientFirstMessage([]byte("n,a=wario,n=mario,r=nonce,iamanextension"))
		if err != nil {
			t.Fatalf("parseClientFirstMessage finished with unexpected error: %s", err)
		}

		expectedGs2CbindFlag := []byte("n")
		expectedAuthzID := []byte("wario")
		expectedUsername := []byte("mario")
		expectedNonce := []byte("nonce")
		expectedGs2Header := []byte("n,a=wario,")
		expectedBare := []byte("n=mario,r=nonce,iamanextension")
		if !bytes.Equal(expectedGs2CbindFlag, message.gs2CbindFlag) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2CbindFlag=%s, but actually finished with gs2CbindFlag=%s", expectedGs2CbindFlag, message.gs2CbindFlag)
		}
		if !bytes.Equal(expectedAuthzID, message.authzID) {
			t.Errorf("parseClientFirstMessage expected to finish with authzID=%s, but actually finished with authzID=%s", expectedAuthzID, message.authzID)
		}
		if !bytes.Equal(expectedUsername, message.username) {
			t.Errorf("parseClientFirstMessage expected to finish with username=%s, but actually finished with username=%s", expectedUsername, message.username)
		}
		if !bytes.Equal(expectedNonce, message.nonce) {
			t.Errorf("parseClientFirstMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
		}
		if !bytes.Equal(expectedGs2Header, message.gs2Header) {
			t.Errorf("parseClientFirstMessage expected to finish with gs2Header=%s, but actually finished with gs2Header=%s", expectedGs2Header, message.gs2Header)
		}
		if !bytes.Equal(expectedBare, message.bare) {
			t.Errorf("parseClientFirstMessage expected to finish with bare=%s, but actually finished with bare=%s", expectedBare, message.bare)
		}
	})

	t.Run("negative_flags", func(t *testing.T) {
		_, err := parseClientFirstMessage([]byte("k,,n=mario,r=nonce"))
		expectedError := `"k" is invalid gs2-cbind-flag`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFirstMessage([]byte("asdf,,n=mario,r=nonce"))
		expectedError = `"asdf" is invalid gs2-cbind-flag`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFirstMessage([]byte("k=k,,n=mario,r=alsononce"))
		expectedError = `"k=k" is invalid gs2-cbind-flag`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFirstMessage([]byte(",,n=mario,r=alsononce"))
		expectedError = `"" is invalid gs2-cbind-flag`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})

	t.Run("negative_message_extensions", func(t *testing.T) {
		_, err := parseClientFirstMessage([]byte("n,,m=ext,n=mario,thisisnonce"))
		expectedError := `SCRAM message extensions are not supported`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})

	t.Run("negative_authzID", func(t *testing.T) {
		_, err := parseClientFirstMessage([]byte("n,b=wario,n=mario,r=nonce"))
		expectedError := `"b=wario" is invalid authzid`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFirstMessage([]byte("y,wario,n=mario,r=alsononce"))
		expectedError = `"wario" is invalid authzid`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})

	t.Run("negative_user", func(t *testing.T) {
		_, err := parseClientFirstMessage([]byte("n,,user=mario,r=nonce"))
		expectedError := `"user=mario" is invalid username`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFirstMessage([]byte("n,,n=,r=nonce"))
		expectedError = `got empty username`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})

	t.Run("negative_nonce", func(t *testing.T) {
		_, err := parseClientFirstMessage([]byte("n,,n=mario,thisisnonce"))
		expectedError := `"thisisnonce" is invalid nonce`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFirstMessage([]byte("y,,n=mario,"))
		expectedError = `"" is invalid nonce`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFirstMessage([]byte("y,,n=mario,r="))
		expectedError = `got empty nonce`
		if err.Error() != expectedError {
			t.Fatalf("parseClientFirstMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})
}

// parseClientFinalMessage test cases
func TestParseClientFinalMessage(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var err error
		c := base64.StdEncoding.EncodeToString([]byte("ab"))
		p := base64.StdEncoding.EncodeToString([]byte("yz"))

		message, err := parseClientFinalMessage([]byte(fmt.Sprintf("c=%s,r=abc,a,s,d,f,p=%s", c, p)))
		if err != nil {
			t.Fatalf("parseClientFinalMessage finished with unexpected error: %s", err)
		}

		expectedChannelBinding := []byte("ab")
		expectedNonce := []byte("abc")
		expectedProof := []byte("yz")
		expectedMessageWithoutProof := []byte(fmt.Sprintf("c=%s,r=abc,a,s,d,f", c))
		if !bytes.Equal(expectedChannelBinding, message.channelBinding) {
			t.Errorf("parseClientFinalMessage expected to finish with channelBinding=%s, but actually finished with channelBinding=%s", expectedChannelBinding, message.channelBinding)
		}
		if !bytes.Equal(expectedNonce, message.nonce) {
			t.Errorf("parseClientFinalMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
		}
		if !bytes.Equal(expectedProof, message.proof) {
			t.Errorf("parseClientFinalMessage expected to finish with proof=%s, but actually finished with proof=%s", expectedProof, message.proof)
		}
		if !bytes.Equal(expectedMessageWithoutProof, message.messageWithoutProof) {
			t.Errorf("parseClientFinalMessage expected to finish with messageWithoutProof=%s, but actually finished with messageWithoutProof=%s", expectedMessageWithoutProof, message.messageWithoutProof)
		}

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("c=%s,r=,p=alphabet", c)))
		expectedError := "got empty nonce (r=...)"
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("c=%s,r=ab,p=", c)))
		expectedError = "got empty proof (p=...)"
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})

	t.Run("empty_values", func(t *testing.T) {
		var err error
		c := base64.StdEncoding.EncodeToString([]byte("ab"))

		_, err = parseClientFinalMessage([]byte("c=,r=abc,a,s,d,f,p=alphabet"))
		expectedError := "got empty channel-binding (c=...)"
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("c=%s,r=,p=alphabet", c)))
		expectedError = "got empty nonce (r=...)"
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("c=%s,r=ab,p=", c)))
		expectedError = "got empty proof (p=...)"
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})

	t.Run("incorrect_order_values", func(t *testing.T) {
		var err error
		c := base64.StdEncoding.EncodeToString([]byte("ab"))

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("c=%s,r=abc,p=alphabet,a=alefbet", c)))
		expectedError := "expected proof (p=...) to be last field"
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("g=h,c=%s,r=abc,p=alphabet,a=alefbet", c)))
		expectedError = "expected channel-binding (c=...) as 1st field, got \"g=h\""
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("g=h,c=%s,r=abc,p=alphabet,a=alefbet", c)))
		expectedError = "expected channel-binding (c=...) as 1st field, got \"g=h\""
		if err.Error() != expectedError {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
		}
	})

	t.Run("incorrect_b64_encoding", func(t *testing.T) {
		var err error
		c := base64.StdEncoding.EncodeToString([]byte("ab"))

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("c=*,r=abc,p=%s", c)))
		expectedPrefix := "cannot decode \"c=*\""
		if !strings.HasPrefix(err.Error(), expectedPrefix) {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error prefix, but actually finished with %q error", expectedPrefix, err)
		}

		_, err = parseClientFinalMessage([]byte(fmt.Sprintf("c=%s,r=abc,p=*", c)))
		expectedPrefix = "cannot decode \"p=*\""
		if !strings.HasPrefix(err.Error(), expectedPrefix) {
			t.Fatalf("parseClientFinalMessage expected to finish with %q error prefix, but actually finished with %q error", expectedPrefix, err)
		}
	})
}

func FuzzClientFinalMessageParsing(f *testing.F) {
	f.Add("cbind", "nonce", "extensions", "proof")
	f.Fuzz(func(t *testing.T, cbind string, nonce string, extensions string, proof string) {
		var err error
		ecbind := base64.StdEncoding.EncodeToString([]byte(cbind))
		enonce := strings.ReplaceAll(nonce, ",", ".")
		eextensions := strings.ReplaceAll(extensions, "p=", "..")
		eproof := base64.StdEncoding.EncodeToString([]byte(proof))

		challenge := []byte(fmt.Sprintf("c=%s,r=%s,%s,p=%s", ecbind, enonce, eextensions, eproof))
		message, err := parseClientFinalMessage(challenge)
		if len(cbind) == 0 {
			expectedError := "got empty channel-binding (c=...)"
			if err.Error() != expectedError {
				t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
			}
		} else if len(nonce) == 0 {
			expectedError := "got empty nonce (r=...)"
			if err.Error() != expectedError {
				t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
			}
		} else if len(proof) == 0 {
			expectedError := "got empty proof (p=...)"
			if err.Error() != expectedError {
				t.Fatalf("parseClientFinalMessage expected to finish with %q error, but actually finished with %q error", expectedError, err)
			}
		} else {
			if err != nil {
				t.Fatalf("parseClientFinalMessage expected to finish without error, but actually finished with %q error", err)
			}
			expectedChannelBinding := []byte(cbind)
			expectedNonce := []byte(enonce)
			expectedProof := []byte(proof)
			if !bytes.Equal(expectedChannelBinding, message.channelBinding) {
				t.Errorf("parseClientFinalMessage expected to finish with channelBinding=%s, but actually finished with channelBinding=%s", expectedChannelBinding, message.channelBinding)
			}
			if !bytes.Equal(expectedNonce, message.nonce) {
				t.Errorf("parseClientFinalMessage expected to finish with nonce=%s, but actually finished with nonce=%s", expectedNonce, message.nonce)
			}
			if !bytes.Equal(expectedProof, message.proof) {
				t.Errorf("parseClientFinalMessage expected to finish with proof=%s, but actually finished with proof=%s", expectedProof, message.proof)
			}
		}
	})
}

// EscapeSaslname test cases
func TestEscapeSaslname(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		expected := []byte("=2C")
		actual := escapeSaslname([]byte(","))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=2Cwario")
		actual = escapeSaslname([]byte("mario,wario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=2C")
		actual = escapeSaslname([]byte("mario,"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2Cwario")
		actual = escapeSaslname([]byte(",wario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=3D")
		actual = escapeSaslname([]byte("="))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=3Dwario")
		actual = escapeSaslname([]byte("mario=wario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=3D")
		actual = escapeSaslname([]byte("mario="))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=3Dwario")
		actual = escapeSaslname([]byte("=wario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2C=3D3D")
		actual = escapeSaslname([]byte(",=3D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2C=3D")
		actual = escapeSaslname([]byte(",="))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2C🙄=3D")
		actual = escapeSaslname([]byte(",🙄="))
		if !bytes.Equal(expected, actual) {
			t.Errorf("EscapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}
	})
}

func FuzzEscapeSaslname(f *testing.F) {
	f.Add("unescaped")
	f.Fuzz(func(t *testing.T, unescaped string) {
		escaped := escapeSaslname([]byte(unescaped))
		escapedFromReplacer := strings.NewReplacer("=", "=3D", ",", "=2C").Replace(unescaped)

		if bytes.Contains(escaped, []byte(",")) {
			t.Errorf("EscapeSaslname result should not contain \",\"")
		}
		if bytes.Contains(escaped, []byte("=")) {
			t.Errorf("EscapeSaslname result should not contain \"=\"")
		}
		expected := []byte(escapedFromReplacer)
		if !bytes.Equal(expected, escaped) {
			t.Errorf("EscapeSaslname result expected to be %q but actually %q", expected, escaped)
		}
	})
}

// UnescapeSaslname test cases

func TestUnescapeSaslname(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		expected := []byte(",")
		actual := unescapeSaslname([]byte("=2C"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario,wario")
		actual = unescapeSaslname([]byte("mario=2Cwario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario,")
		actual = unescapeSaslname([]byte("mario=2C"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte(",wario")
		actual = unescapeSaslname([]byte("=2Cwario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=")
		actual = unescapeSaslname([]byte("=3D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=")
		actual = unescapeSaslname([]byte("=3D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=wario")
		actual = unescapeSaslname([]byte("mario=3Dwario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=")
		actual = unescapeSaslname([]byte("mario=3D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=wario")
		actual = unescapeSaslname([]byte("=3Dwario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte(",=")
		actual = unescapeSaslname([]byte("=2C=3D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2C")
		actual = unescapeSaslname([]byte("=3D2C"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte(",3D")
		actual = unescapeSaslname([]byte("=2C3D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2D")
		actual = unescapeSaslname([]byte("=2D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=3C")
		actual = unescapeSaslname([]byte("=3C"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2c")
		actual = unescapeSaslname([]byte("=2c"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=3d")
		actual = unescapeSaslname([]byte("=3d"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=2")
		actual = unescapeSaslname([]byte("=2"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("=3")
		actual = unescapeSaslname([]byte("=3"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=wario")
		actual = unescapeSaslname([]byte("mario=wario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte("mario=2wario")
		actual = unescapeSaslname([]byte("mario=2wario"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}

		expected = []byte(",🙄=")
		actual = unescapeSaslname([]byte("=2C🙄=3D"))
		if !bytes.Equal(expected, actual) {
			t.Errorf("UnescapeSaslname expected to return %s, but actually returned %s", expected, actual)
		}
	})
}

func FuzzUnescapeSaslname(f *testing.F) {
	f.Add("escaped")
	f.Fuzz(func(t *testing.T, escaped string) {
		unescaped := unescapeSaslname([]byte(escaped))
		unescapedFromReplacer := strings.NewReplacer("=3D", "=", "=2C", ",").Replace(escaped)

		expected := []byte(unescapedFromReplacer)
		if !bytes.Equal(expected, unescaped) {
			t.Errorf("UnescapeSaslname result expected to be %q but actually %q", expected, unescaped)
		}
	})
}
