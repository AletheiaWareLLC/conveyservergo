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
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/conveyservergo"
	"github.com/AletheiaWareLLC/testinggo"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeAccountTemplate(t *testing.T) *template.Template {
	t.Helper()
	tmplt, err := template.New("").Parse(`{{ .Alias }}{{ .Balance }}`)
	testinggo.AssertNoError(t, err)
	return tmplt
}

func TestAccountHandler(t *testing.T) {
	alias := "Alice"
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error("Could not generate key:", err)
	}
	t.Run("GETSignedIn", func(t *testing.T) {
		// Show Account Info
		sessionstore := main.NewMemorySessionStore()
		session, err := sessionstore.CreateSignInSession(alias, key)
		testinggo.AssertNoError(t, err)

		ledger := conveygo.NewLedger(&bcgo.Node{})
		ledger.Earned[alias] = 1234

		request := makeGetAccountRequest(t)
		request.AddCookie(main.CreateSignInSessionCookie(session, time.Hour))
		response := httptest.NewRecorder()

		handler := main.AccountHandler(sessionstore, ledger, makeAccountTemplate(t))
		handler(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
		}

		actual := response.Body.String()
		expected := "Alice1234"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("GETNotSignedIn", func(t *testing.T) {
		// Redirect to Sign In page
		sessionstore := main.NewMemorySessionStore()

		ledger := conveygo.NewLedger(&bcgo.Node{})

		request := makeGetAccountRequest(t)
		response := httptest.NewRecorder()

		handler := main.AccountHandler(sessionstore, ledger, makeAccountTemplate(t))
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
}

func makeGetAccountRequest(t *testing.T) *http.Request {
	request, err := http.NewRequest(http.MethodGet, "/account", nil)
	testinggo.AssertNoError(t, err)
	return request
}
