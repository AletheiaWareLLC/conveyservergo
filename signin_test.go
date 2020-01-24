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
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/conveyservergo"
	"github.com/AletheiaWareLLC/testinggo"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func makeSignInTemplate(t *testing.T) *template.Template {
	t.Helper()
	tmplt, err := template.New("").Parse("Sign In")
	testinggo.AssertNoError(t, err)
	return tmplt
}

func TestSignInHandler(t *testing.T) {
	alias := "Alice"
	password := "password1234"
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error("Could not generate key:", err)
	}
	t.Run("GETSignedIn", func(t *testing.T) {
		// Redirect to Account page
		sessionstore := main.NewMemorySessionStore()
		session, err := sessionstore.CreateSignInSession(alias, key)
		testinggo.AssertNoError(t, err)
		userstore := conveygo.NewMemoryStore()

		request := makeGetSignInRequest(t)
		request.AddCookie(main.CreateSignInSessionCookie(session, time.Hour))
		response := httptest.NewRecorder()

		handler := main.SignInHandler(sessionstore, userstore, makeSignInTemplate(t))
		handler(response, request)

		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Body.String()
		expected := `<a href="/account">Found</a>.

`

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("GETNotSignedIn", func(t *testing.T) {
		// Show Sign In form
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()

		request := makeGetSignInRequest(t)
		response := httptest.NewRecorder()

		handler := main.SignInHandler(sessionstore, userstore, makeSignInTemplate(t))
		handler(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
		}

		actual := response.Body.String()
		expected := "Sign In"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("POSTSignedIn", func(t *testing.T) {
		// Redirect to Account page
		sessionstore := main.NewMemorySessionStore()
		session, err := sessionstore.CreateSignInSession(alias, key)
		testinggo.AssertNoError(t, err)
		userstore := conveygo.NewMemoryStore()

		request := makePostSignInRequest(t)
		request.AddCookie(main.CreateSignInSessionCookie(session, time.Hour))
		response := httptest.NewRecorder()

		handler := main.SignInHandler(sessionstore, userstore, makeSignInTemplate(t))
		handler(response, request)

		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Header().Get("Location")
		expected := "/account"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("POSTNotSignedIn", func(t *testing.T) {
		// Sign User In and Redirect to Account page
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		testinggo.AssertNoError(t, userstore.AddKey(alias, []byte(password), key))

		data := url.Values{}
		data.Set("alias", alias)
		data.Set("password", password)
		request, err := http.NewRequest(http.MethodPost, "/sign-in", strings.NewReader(data.Encode()))
		testinggo.AssertNoError(t, err)
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()

		handler := main.SignInHandler(sessionstore, userstore, makeSignInTemplate(t))
		handler(response, request)

		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Header().Get("Location")
		expected := "/account"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}

		cookies := response.Result().Cookies()
		if len(cookies) != 1 {
			t.Error("Missing Session Cookie")
		}

		cookie := cookies[0]
		if cookie.Name != main.SESSION_COOKIE_SIGN_IN {
			t.Errorf("Wrong cookie; expected '%s', got '%s'", main.SESSION_COOKIE_SIGN_IN, cookie.Name)
		}

		if !sessionstore.IsValidSignInSession(cookie.Value) {
			t.Error("User was not signed in")
		}
	})
}

func makeGetSignInRequest(t *testing.T) *http.Request {
	request, err := http.NewRequest(http.MethodGet, "/sign-in", nil)
	testinggo.AssertNoError(t, err)
	return request
}

func makePostSignInRequest(t *testing.T) *http.Request {
	request, err := http.NewRequest(http.MethodPost, "/sign-in", nil)
	testinggo.AssertNoError(t, err)
	return request
}
