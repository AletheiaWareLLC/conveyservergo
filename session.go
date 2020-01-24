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

package main

import (
	"crypto/rsa"
	"errors"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/cryptogo"
	"net/http"
	"regexp"
	"time"
)

const (
	ERROR_INVALID_EMAIL          = "Invalid Email Address"
	ERROR_INVALID_NAME           = "Invalid Name"
	ERROR_LEGALESE_REQUIRED      = "You Must Read, Understand, and Agree to the Legalese"
	ERROR_PASSWORD_TOO_SHORT     = "Password Too Short"
	ERROR_PASSWORDS_DO_NOT_MATCH = "Passwords Do Not Match"
	MINIMUM_PASSWORD_LENGTH      = 12
	SESSION_COOKIE_SIGN_IN       = "sign-in-session"
	SESSION_COOKIE_SIGN_UP       = "sign-up-session"
	SESSION_ID_LENGTH            = 16
	SESSION_TIMEOUT_SIGN_IN      = 30 * time.Minute
	SESSION_TIMEOUT_SIGN_UP      = 10 * time.Minute
)

// This is not intended to validate every possible email address, instead a verification code will be sent to ensure the email works
var emails = regexp.MustCompile(`^.+@.+$`)

func ValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return emails.MatchString(email)
}

func ValidPassword(password string) bool {
	return len(password) >= MINIMUM_PASSWORD_LENGTH
}

func MatchingPasswords(password, confirmation string) bool {
	return password == confirmation
}

type SignUpSession struct {
	Error         string
	Legalese      string
	Name          string
	Email         string
	Challenge     string
	Verification  string
	Alias         string
	Password      string
	Confirmation  string
	PaymentMethod string
}

func (s *SignUpSession) Validate() error {
	// Check legalese accepted
	if s.Legalese != "accept" {
		return errors.New(ERROR_LEGALESE_REQUIRED)
	}
	// Check valid name
	if len(s.Name) == 0 {
		return errors.New(ERROR_INVALID_NAME)
	}
	// Check valid email
	if !ValidEmail(s.Email) {
		return errors.New(ERROR_INVALID_EMAIL)
	}
	// Check valid password and matching confirm
	if !ValidPassword(s.Password) {
		return errors.New(ERROR_PASSWORD_TOO_SHORT)
	}
	if !MatchingPasswords(s.Password, s.Confirmation) {
		return errors.New(ERROR_PASSWORDS_DO_NOT_MATCH)
	}
	return nil
}

type SignInSession struct {
	Alias             string
	Key               *rsa.PrivateKey
	AddPaymentMethod  *AddPaymentMethodSession
	DraftContribution *DraftContributionSession
	TokenPurchase     *TokenPurchaseSession
	TokenTransfer     *TokenTransferSession
}

type AddPaymentMethodSession struct {
	Error         string
	PaymentMethod string
}

type DraftContributionSession struct {
	Conversation       *conveygo.Conversation
	ConversationHash   []byte
	ConversationRecord *bcgo.Record
	ConversationCost   uint64
	Message            *conveygo.Message
	MessageHash        []byte
	MessageRecord      *bcgo.Record
	MessageCost        uint64
}

type TokenPurchaseSession struct {
	Error   string
	Product map[string]*Product
}

type TokenTransferSession struct {
	Error string
}

type SessionStore interface {
	// Sign Up
	GetSignUpSessionTimeout() time.Duration
	SetSignUpSessionTimeout(time.Duration)
	CreateSignUpSession() (string, error)
	GetSignUpSession(id string) *SignUpSession

	// Sign In
	GetSignInSessionTimeout() time.Duration
	SetSignInSessionTimeout(time.Duration)
	CreateSignInSession(alias string, key *rsa.PrivateKey) (string, error)
	GetSignInSession(id string) *SignInSession
	RefreshSignInSession(session *SignInSession) (string, error)
	IsValidSignInSession(id string) bool
	DeleteSignInSession(id string)
}

func CreateSessionId() (string, error) {
	return cryptogo.RandomString(SESSION_ID_LENGTH)
}

func CreateSignInSessionCookie(session string, timeout time.Duration) *http.Cookie {
	return CreateCookie(SESSION_COOKIE_SIGN_IN, session, timeout)
}

func CreateSignUpSessionCookie(session string, timeout time.Duration) *http.Cookie {
	return CreateCookie(SESSION_COOKIE_SIGN_UP, session, timeout)
}

func GetSignInSessionCookie(r *http.Request) (*http.Cookie, error) {
	return GetCookie(SESSION_COOKIE_SIGN_IN, r)
}

func GetSignUpSessionCookie(r *http.Request) (*http.Cookie, error) {
	return GetCookie(SESSION_COOKIE_SIGN_UP, r)
}

func CreateCookie(name, value string, timeout time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(timeout),
		Secure:   !bcgo.IsDebug(),
		HttpOnly: true,
	}
}

func GetCookie(name string, r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}
