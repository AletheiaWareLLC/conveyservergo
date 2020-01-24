/*
 * Copyright 2020 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main_test

import (
	"crypto/rsa"
	"encoding/base64"
	"github.com/AletheiaWareLLC/conveyservergo"
	"github.com/AletheiaWareLLC/testinggo"
	"testing"
	"time"
)

func testSessionStore_CreateSignUpSession(t *testing.T, s main.SessionStore) {
	t.Helper()
	id, err := s.CreateSignUpSession()
	testinggo.AssertNoError(t, err)
	bytes, err := base64.RawURLEncoding.DecodeString(id)
	testinggo.AssertNoError(t, err)
	if len(bytes) != 16 {
		t.Error("ID is not 16 bytes")
	}
}

func testSessionStore_CreateSignUpSession_Expiry(t *testing.T, s main.SessionStore) {
	t.Helper()
	s.SetSignUpSessionTimeout(time.Second)
	id, err := s.CreateSignUpSession()
	testinggo.AssertNoError(t, err)
	time.Sleep(time.Second * 2)
	session := s.GetSignUpSession(id)
	if session != nil {
		t.Error("Sign Up Session did not expire")
	}
}

func testSessionStore_GetSignUpSession_Exists(t *testing.T, s main.SessionStore) {
	t.Helper()
	id, err := s.CreateSignUpSession()
	testinggo.AssertNoError(t, err)
	session := s.GetSignUpSession(id)
	if session == nil {
		t.Error("Sign Up Session should not be nil")
	}
}

func testSessionStore_GetSignUpSession_NotExists(t *testing.T, s main.SessionStore) {
	t.Helper()
	session := s.GetSignUpSession("DoesNotExist")
	if session != nil {
		t.Error("Sign Up Session should be nil")
	}
}

func testSessionStore_CreateSignInSession(t *testing.T, s main.SessionStore, alias string, key *rsa.PrivateKey) {
	t.Helper()
	id, err := s.CreateSignInSession(alias, key)
	testinggo.AssertNoError(t, err)
	bytes, err := base64.RawURLEncoding.DecodeString(id)
	testinggo.AssertNoError(t, err)
	if len(bytes) != 16 {
		t.Error("ID is not 16 bytes")
	}
}

func testSessionStore_CreateSignInSession_Expiry(t *testing.T, s main.SessionStore, alias string, key *rsa.PrivateKey) {
	t.Helper()
	s.SetSignInSessionTimeout(time.Second)
	id, err := s.CreateSignInSession(alias, key)
	testinggo.AssertNoError(t, err)
	time.Sleep(time.Second * 2)
	session := s.GetSignInSession(id)
	if session != nil {
		t.Error("Sign In Session did not expire")
	}
}

func testSessionStore_GetSignInSession_Exists(t *testing.T, s main.SessionStore, alias string, key *rsa.PrivateKey) {
	t.Helper()
	id, err := s.CreateSignInSession(alias, key)
	testinggo.AssertNoError(t, err)
	session := s.GetSignInSession(id)
	if session == nil {
		t.Error("Sign In Session should not be nil")
	}
	if alias != session.Alias {
		t.Errorf("Incorrect alias; expected '%s', got '%s'", alias, session.Alias)
	}
}

func testSessionStore_GetSignInSession_NotExists(t *testing.T, s main.SessionStore) {
	t.Helper()
	session := s.GetSignInSession("DoesNotExist")
	if session != nil {
		t.Error("Sign In Session should be nil")
	}
}

func testSessionStore_RefreshSignInSession(t *testing.T, s main.SessionStore, alias string, key *rsa.PrivateKey) {
	t.Helper()
	id, err := s.RefreshSignInSession(&main.SignInSession{})
	testinggo.AssertNoError(t, err)
	bytes, err := base64.RawURLEncoding.DecodeString(id)
	testinggo.AssertNoError(t, err)
	if len(bytes) != 16 {
		t.Error("ID is not 16 bytes")
	}
}

func testSessionStore_RefreshSignInSession_Expiry(t *testing.T, s main.SessionStore, alias string, key *rsa.PrivateKey) {
	t.Helper()
	s.SetSignInSessionTimeout(time.Second)
	id, err := s.RefreshSignInSession(&main.SignInSession{})
	testinggo.AssertNoError(t, err)
	time.Sleep(time.Second * 2)
	session := s.GetSignInSession(id)
	if session != nil {
		t.Error("Sign In Session did not expire")
	}
}

func testSessionStore_IsValidSignInSession_Valid(t *testing.T, s main.SessionStore, alias string, key *rsa.PrivateKey) {
	t.Helper()
	id, err := s.CreateSignInSession(alias, key)
	testinggo.AssertNoError(t, err)
	if !s.IsValidSignInSession(id) {
		t.Error("Sign In Session should be valid")
	}
}

func testSessionStore_IsValidSignInSession_Invalid(t *testing.T, s main.SessionStore) {
	t.Helper()
	if s.IsValidSignInSession("DoesNotExist") {
		t.Error("Sign In Session should not be valid")
	}
}

func testSessionStore_DeleteSignInSession_Exists(t *testing.T, s main.SessionStore, alias string, key *rsa.PrivateKey) {
	t.Helper()
	id, err := s.CreateSignInSession(alias, key)
	testinggo.AssertNoError(t, err)
	s.DeleteSignInSession(id)
	session := s.GetSignInSession(id)
	if session != nil {
		t.Error("Sign In Session should be nil")
	}
}

func testSessionStore_DeleteSignInSession_NotExists(t *testing.T, s main.SessionStore) {
	t.Helper()
	s.DeleteSignInSession("DoesNotExist")
}
