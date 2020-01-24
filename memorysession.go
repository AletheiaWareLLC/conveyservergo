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
	// "log"
	"time"
)

type MemorySessionStore struct {
	SignInTimeout time.Duration
	SignUpTimeout time.Duration
	SignIns       map[string]*SignInSession
	SignUps       map[string]*SignUpSession
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		SignInTimeout: SESSION_TIMEOUT_SIGN_IN,
		SignUpTimeout: SESSION_TIMEOUT_SIGN_UP,
		SignIns:       make(map[string]*SignInSession),
		SignUps:       make(map[string]*SignUpSession),
	}
}

func (s *MemorySessionStore) GetSignInSessionTimeout() time.Duration {
	return s.SignInTimeout
}

func (s *MemorySessionStore) GetSignUpSessionTimeout() time.Duration {
	return s.SignUpTimeout
}

func (s *MemorySessionStore) SetSignInSessionTimeout(timeout time.Duration) {
	s.SignInTimeout = timeout
}

func (s *MemorySessionStore) SetSignUpSessionTimeout(timeout time.Duration) {
	s.SignUpTimeout = timeout
}

func (s *MemorySessionStore) CreateSignUpSession() (string, error) {
	id, err := CreateSessionId()
	if err != nil {
		return "", err
	}
	// log.Println("Creating Sign Up Session", id)

	s.SignUps[id] = &SignUpSession{}

	go func() {
		// Delete session after timeout
		time.Sleep(s.SignUpTimeout)
		// log.Println("Expiring Sign Up Session", id)
		delete(s.SignUps, id)
	}()

	return id, nil
}

func (s *MemorySessionStore) GetSignUpSession(id string) *SignUpSession {
	return s.SignUps[id]
}

func (s *MemorySessionStore) CreateSignInSession(alias string, key *rsa.PrivateKey) (string, error) {
	return s.RefreshSignInSession(&SignInSession{
		Alias: alias,
		Key:   key,
	})
}

func (s *MemorySessionStore) GetSignInSession(id string) *SignInSession {
	return s.SignIns[id]
}

func (s *MemorySessionStore) RefreshSignInSession(session *SignInSession) (string, error) {
	// TODO(v2) ensure session is not nil
	id, err := CreateSessionId()
	if err != nil {
		return "", err
	}
	// log.Println("Creating Sign In Session", id)

	s.SignIns[id] = session

	go func() {
		// Delete session after timeout
		time.Sleep(s.SignInTimeout)
		// log.Println("Expiring Sign In Session", id)
		delete(s.SignIns, id)
	}()

	return id, nil
}

func (s *MemorySessionStore) IsValidSignInSession(id string) bool {
	_, ok := s.SignIns[id]
	return ok
}

func (s *MemorySessionStore) DeleteSignInSession(id string) {
	delete(s.SignIns, id)
}
