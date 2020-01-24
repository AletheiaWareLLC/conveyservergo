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
	"crypto/rand"
	"crypto/rsa"
	"github.com/AletheiaWareLLC/conveyservergo"
	"testing"
)

func TestMemorySessionStore(t *testing.T) {
	alias := "Alice"
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error("Could not generate key:", err)
	}
	t.Run("CreateSignUpSession", func(t *testing.T) {
		testSessionStore_CreateSignUpSession(t, main.NewMemorySessionStore())
		t.Run("Expiry", func(t *testing.T) {
			testSessionStore_CreateSignUpSession_Expiry(t, main.NewMemorySessionStore())
		})
	})
	t.Run("GetSignUpSession", func(t *testing.T) {
		t.Run("Exists", func(t *testing.T) {
			testSessionStore_GetSignUpSession_Exists(t, main.NewMemorySessionStore())
		})
		t.Run("NotExists", func(t *testing.T) {
			testSessionStore_GetSignUpSession_NotExists(t, main.NewMemorySessionStore())
		})
	})
	t.Run("CreateSignInSession", func(t *testing.T) {
		testSessionStore_CreateSignInSession(t, main.NewMemorySessionStore(), alias, key)
		t.Run("Expiry", func(t *testing.T) {
			testSessionStore_CreateSignInSession_Expiry(t, main.NewMemorySessionStore(), alias, key)
		})
	})
	t.Run("GetSignInSession", func(t *testing.T) {
		t.Run("Exists", func(t *testing.T) {
			testSessionStore_GetSignInSession_Exists(t, main.NewMemorySessionStore(), alias, key)
		})
		t.Run("NotExists", func(t *testing.T) {
			testSessionStore_GetSignInSession_NotExists(t, main.NewMemorySessionStore())
		})
	})
	t.Run("RefreshSignInSession", func(t *testing.T) {
		testSessionStore_RefreshSignInSession(t, main.NewMemorySessionStore(), alias, key)
		t.Run("Expiry", func(t *testing.T) {
			testSessionStore_RefreshSignInSession_Expiry(t, main.NewMemorySessionStore(), alias, key)
		})
	})
	t.Run("IsValidSignInSession", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			testSessionStore_IsValidSignInSession_Valid(t, main.NewMemorySessionStore(), alias, key)
		})
		t.Run("Invalid", func(t *testing.T) {
			testSessionStore_IsValidSignInSession_Invalid(t, main.NewMemorySessionStore())
		})
	})
	t.Run("DeleteSignInSession", func(t *testing.T) {
		t.Run("Exists", func(t *testing.T) {
			testSessionStore_DeleteSignInSession_Exists(t, main.NewMemorySessionStore(), alias, key)
		})
		t.Run("NotExists", func(t *testing.T) {
			testSessionStore_DeleteSignInSession_NotExists(t, main.NewMemorySessionStore())
		})
	})
}
