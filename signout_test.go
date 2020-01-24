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
	"github.com/AletheiaWareLLC/testinggo"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeSignOutTemplate(t *testing.T) *template.Template {
	t.Helper()
	tmplt, err := template.New("").Parse("Sign Out")
	testinggo.AssertNoError(t, err)
	return tmplt
}

func TestSignOutHandler(t *testing.T) {
	alias := "Alice"
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error("Could not generate key:", err)
	}
	t.Run("GETSignedIn", func(t *testing.T) {
		// Show Sign Out button
		sessionstore := main.NewMemorySessionStore()
		session, err := sessionstore.CreateSignInSession(alias, key)
		testinggo.AssertNoError(t, err)

		request := makeGetSignOutRequest(t)
		request.AddCookie(main.CreateSignInSessionCookie(session, time.Hour))
		response := httptest.NewRecorder()

		handler := main.SignOutHandler(sessionstore, makeSignOutTemplate(t))
		handler(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
		}

		actual := response.Body.String()
		expected := "Sign Out"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("GETNotSignedIn", func(t *testing.T) {
		// Redirect to Sign In page
		sessionstore := main.NewMemorySessionStore()

		request := makeGetSignOutRequest(t)
		response := httptest.NewRecorder()

		handler := main.SignOutHandler(sessionstore, makeSignOutTemplate(t))
		handler(response, request)

		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Body.String()
		expected := `<a href="/sign-in">Found</a>.

`

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("POSTSignedIn", func(t *testing.T) {
		// Sign User Out and Redirect to Home page
		sessionstore := main.NewMemorySessionStore()
		session, err := sessionstore.CreateSignInSession(alias, key)
		testinggo.AssertNoError(t, err)

		request := makePostSignOutRequest(t)
		request.AddCookie(main.CreateSignInSessionCookie(session, time.Hour))
		response := httptest.NewRecorder()

		handler := main.SignOutHandler(sessionstore, makeSignOutTemplate(t))
		handler(response, request)

		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Header().Get("Location")
		expected := "/"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}

		if sessionstore.IsValidSignInSession(session) {
			t.Error("User was not signed out")
		}
	})
	t.Run("POSTNotSignedIn", func(t *testing.T) {
		// Redirect to Sign In page
		sessionstore := main.NewMemorySessionStore()

		request := makePostSignOutRequest(t)
		response := httptest.NewRecorder()

		handler := main.SignOutHandler(sessionstore, makeSignOutTemplate(t))
		handler(response, request)

		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Header().Get("Location")
		expected := "/sign-in"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
}

func makeGetSignOutRequest(t *testing.T) *http.Request {
	request, err := http.NewRequest(http.MethodGet, "/sign-out", nil)
	testinggo.AssertNoError(t, err)
	return request
}

func makePostSignOutRequest(t *testing.T) *http.Request {
	request, err := http.NewRequest(http.MethodPost, "/sign-out", nil)
	testinggo.AssertNoError(t, err)
	return request
}
