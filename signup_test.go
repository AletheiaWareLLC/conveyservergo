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
	"fmt"
	"github.com/AletheiaWareLLC/aliasgo"
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

func makeSignUpTemplate(t *testing.T) *template.Template {
	t.Helper()
	tmplt, err := template.New("").Parse("{{ if .Error}}{{ .Error }}{{ else }}Sign Up{{ end }}")
	testinggo.AssertNoError(t, err)
	return tmplt
}

func TestSignUpHandler(t *testing.T) {
	email := "alice@example.com"
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
		emailverifier := makeMockEmailVerifier(t, "test1234")

		request := makeGetSignUpRequest(t)
		request.AddCookie(main.CreateSignInSessionCookie(session, time.Hour))
		response := httptest.NewRecorder()

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))
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
		// Show Sign Up form
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		request := makeGetSignUpRequest(t)
		response := httptest.NewRecorder()

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))
		handler(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
		}

		actual := response.Body.String()
		expected := "Sign Up"

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
		emailverifier := makeMockEmailVerifier(t, "test1234")

		request := makePostSignUpRequest(t)
		request.AddCookie(main.CreateSignInSessionCookie(session, time.Hour))
		response := httptest.NewRecorder()

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))
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
	t.Run("POSTWithoutGET", func(t *testing.T) {
		// Redirect to sign up
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		request := makePostSignUpRequest(t)
		response := httptest.NewRecorder()

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))
		handler(response, request)

		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Header().Get("Location")
		expected := "/sign-up"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("Legalese", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))

		cookie := getSignUpCookie(t, handler)

		data := &url.Values{}
		request := makePostSignUpRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		checkPostResponse(t, response)
		checkGetResponse(t, handler, cookie, main.ERROR_LEGALESE_REQUIRED)
	})
	t.Run("Name", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))

		cookie := getSignUpCookie(t, handler)

		data := &url.Values{}
		data.Set("legalese", "accept")
		request := makePostSignUpRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		checkPostResponse(t, response)
		checkGetResponse(t, handler, cookie, main.ERROR_INVALID_NAME)
	})
	t.Run("Email", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))

		cookie := getSignUpCookie(t, handler)

		data := &url.Values{}
		data.Set("legalese", "accept")
		data.Set("name", alias)
		data.Set("email", "notanemail")
		request := makePostSignUpRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		checkPostResponse(t, response)
		checkGetResponse(t, handler, cookie, main.ERROR_INVALID_EMAIL)
	})
	t.Run("Password_Length", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))

		cookie := getSignUpCookie(t, handler)

		data := &url.Values{}
		data.Set("legalese", "accept")
		data.Set("name", alias)
		data.Set("email", email)
		data.Set("password", "password")
		request := makePostSignUpRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		checkPostResponse(t, response)
		checkGetResponse(t, handler, cookie, main.ERROR_PASSWORD_TOO_SHORT)
	})
	t.Run("Password_Match", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))

		cookie := getSignUpCookie(t, handler)

		data := &url.Values{}
		data.Set("legalese", "accept")
		data.Set("name", alias)
		data.Set("email", email)
		data.Set("password", password)
		data.Set("confirmation", "password4321")
		request := makePostSignUpRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		checkPostResponse(t, response)
		checkGetResponse(t, handler, cookie, main.ERROR_PASSWORDS_DO_NOT_MATCH)
	})
	t.Run("Alias", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		testinggo.AssertNoError(t, userstore.AddKey(alias, []byte(password), key))
		emailverifier := makeMockEmailVerifier(t, "test1234")

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))

		cookie := getSignUpCookie(t, handler)

		data := &url.Values{}
		data.Set("legalese", "accept")
		data.Set("name", alias)
		data.Set("email", email)
		data.Set("password", password)
		data.Set("confirmation", password)
		data.Set("alias", alias)
		request := makePostSignUpRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		checkPostResponse(t, response)
		checkGetResponse(t, handler, cookie, fmt.Sprintf(aliasgo.ERROR_ALIAS_ALREADY_REGISTERED, alias))
	})
	t.Run("POSTNotSignedIn", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		emailverifier := makeMockEmailVerifier(t, "test1234")

		handler := main.SignUpHandler(sessionstore, userstore, emailverifier, makeSignUpTemplate(t))

		cookie := getSignUpCookie(t, handler)

		data := &url.Values{}
		data.Set("legalese", "accept")
		data.Set("name", alias)
		data.Set("alias", alias)
		data.Set("password", password)
		data.Set("confirmation", password)
		data.Set("email", email)
		request := makePostSignUpRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Header().Get("Location")
		expected := "/sign-up-verification"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}

		if emailverifier.Email != email {
			t.Errorf("Wrong email; expected '%s', got '%s'", emailverifier.Email, email)
		}
	})
}

func TestSignUpVerificationHandler(t *testing.T) {
	email := "alice@example.com"
	alias := "Alice"
	password := "password1234"
	t.Run("GETSignedIn", func(t *testing.T) {
		// Redirect to Account page
		// TODO(v1)
	})
	t.Run("GETNotSignedIn", func(t *testing.T) {
		// Show Sign Up Verification form
		// TODO(v1)
	})
	t.Run("POSTSignedIn", func(t *testing.T) {
		// Redirect to Account page
		// TODO(v1)
	})
	t.Run("POSTNotSignedIn", func(t *testing.T) {
		sessionstore := main.NewMemorySessionStore()
		userstore := conveygo.NewMemoryStore()
		paymentprocessor := makeMockPaymentProcessor(t)
		emailwelcomer := makeMockEmailWelcomer(t)

		id, err := sessionstore.CreateSignUpSession()
		testinggo.AssertNoError(t, err)
		session := sessionstore.GetSignUpSession(id)
		session.Email = email
		session.Alias = alias
		session.Password = password
		session.Name = alias
		session.Challenge = "challenge1234"
		cookie := main.CreateSignUpSessionCookie(id, sessionstore.GetSignUpSessionTimeout())

		handler := main.SignUpVerificationHandler(sessionstore, userstore, paymentprocessor, emailwelcomer, makeSignUpTemplate(t))

		data := &url.Values{}
		data.Set("verification", "challenge1234")
		request := makePostSignUpVerificationRequestForm(t, data)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		handler(response, request)
		if response.Code != http.StatusFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
		}

		actual := response.Header().Get("Location")
		expected := "/signed-up.html"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}

		if !userstore.HasKey(alias) {
			t.Error("User was not added")
		}

		if emailwelcomer.Alias != alias {
			t.Errorf("Wrong alias; expected '%s', got '%s'", emailwelcomer.Alias, alias)
		}

		if emailwelcomer.Email != email {
			t.Errorf("Wrong email; expected '%s', got '%s'", emailwelcomer.Email, email)
		}
	})
}

func getSignUpCookie(t *testing.T, handler func(http.ResponseWriter, *http.Request)) *http.Cookie {
	t.Helper()
	response := httptest.NewRecorder()
	handler(response, makeGetSignUpRequest(t))

	if response.Code != http.StatusOK {
		t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
	}

	actual := response.Body.String()
	expected := "Sign Up"

	if actual != expected {
		t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
	}

	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Error("Missing Session Cookie")
	}

	cookie := cookies[0]
	if cookie.Name != main.SESSION_COOKIE_SIGN_UP {
		t.Errorf("Wrong cookie; expected '%s', got '%s'", main.SESSION_COOKIE_SIGN_UP, cookie.Name)
	}
	return cookie
}

func checkPostResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Code != http.StatusFound {
		t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusFound, response.Code)
	}

	actual := response.Header().Get("Location")
	expected := "/sign-up"

	if actual != expected {
		t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
	}
}

func checkGetResponse(t *testing.T, handler func(http.ResponseWriter, *http.Request), cookie *http.Cookie, expected string) {
	t.Helper()
	request := makeGetSignUpRequest(t)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
	}

	actual := response.Body.String()

	if actual != expected {
		t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
	}
}

func makeGetSignUpRequest(t *testing.T) *http.Request {
	t.Helper()
	return makeGetRequest(t, "/sign-up")
}

func makePostSignUpRequest(t *testing.T) *http.Request {
	t.Helper()
	return makePostRequest(t, "/sign-up")
}

func makePostSignUpRequestForm(t *testing.T, data *url.Values) *http.Request {
	t.Helper()
	return makePostRequestForm(t, "/sign-up", data)
}

func makeGetSignUpVerificationRequest(t *testing.T) *http.Request {
	t.Helper()
	return makeGetRequest(t, "/sign-up-verification")
}

func makePostSignUpVerificationRequest(t *testing.T) *http.Request {
	t.Helper()
	return makePostRequest(t, "/sign-up-verification")
}

func makePostSignUpVerificationRequestForm(t *testing.T, data *url.Values) *http.Request {
	t.Helper()
	return makePostRequestForm(t, "/sign-up-verification", data)
}

func makeGetRequest(t *testing.T, page string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, page, nil)
	testinggo.AssertNoError(t, err)
	return request
}

func makePostRequest(t *testing.T, page string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, page, nil)
	testinggo.AssertNoError(t, err)
	return request
}

func makePostRequestForm(t *testing.T, page string, data *url.Values) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, page, strings.NewReader(data.Encode()))
	testinggo.AssertNoError(t, err)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return request
}
