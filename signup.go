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
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/AletheiaWareLLC/aliasgo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/conveygo"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type SignUpTemplate struct {
	Error        string
	Legalese     string
	Beta         bool
	Name         string
	Email        string
	Alias        string
	Password     string
	Confirmation string
}

func SignUpHandler(sessions SessionStore, users conveygo.UserStore, verifier EmailVerifier, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path, r.Header)
		cookie, err := GetSignInSessionCookie(r)
		if err == nil && sessions.IsValidSignInSession(cookie.Value) {
			RedirectAccount(w, r)
			return
		}
		session := ""
		cookie, err = GetSignUpSessionCookie(r)
		if err == nil {
			session = cookie.Value
		}
		switch r.Method {
		case "GET":
			// Show sign-up page
			var s *SignUpSession
			for {
				s = sessions.GetSignUpSession(session)
				if s != nil {
					break
				}
				id, err := sessions.CreateSignUpSession()
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				} else {
					session = id
					http.SetCookie(w, CreateSignUpSessionCookie(session, sessions.GetSignUpSessionTimeout()))
				}
			}
			data := &SignUpTemplate{
				Error:        s.Error,
				Legalese:     s.Legalese,
				Beta:         bcgo.IsBeta(),
				Name:         s.Name,
				Email:        s.Email,
				Alias:        s.Alias,
				Password:     s.Password,
				Confirmation: s.Confirmation,
			}
			if err := template.Execute(w, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			}
			return
		case "POST":
			// Try sign user up
			s := sessions.GetSignUpSession(session)
			if s == nil {
				http.SetCookie(w, CreateSignUpSessionCookie("", sessions.GetSignUpSessionTimeout()))
				RedirectSignUp(w, r)
				return
			} else {
				s.Error = ""
				s.Legalese = r.FormValue("legalese")
				s.Name = r.FormValue("name")
				s.Email = r.FormValue("email")
				s.Alias = strings.Join(strings.Fields(r.FormValue("alias")), "") // Strip whitespace
				s.Password = r.FormValue("password")
				s.Confirmation = r.FormValue("confirmation")
				log.Println(r.Form) // TODO(v1) remove
			}
			err := s.Validate()
			if err != nil {
				log.Println(err)
				s.Error = err.Error()
			} else {
				// Check valid alias
				if err := aliasgo.ValidateAlias(s.Alias); err != nil {
					log.Println(err)
					s.Error = err.Error()
				} else {
					// Check unique alias
					if users.HasKey(s.Alias) {
						s.Error = fmt.Sprintf(aliasgo.ERROR_ALIAS_ALREADY_REGISTERED, s.Alias)
					} else {
						if verifier == nil {
							log.Println("Skipping Email Verification")
							s.Challenge = "skip"
							s.Verification = "skip"
							RedirectSignUpVerification(w, r)
							return
						} else {
							code, err := verifier.VerifyEmail(s.Email)
							if err != nil {
								log.Println(err)
								s.Error = err.Error()
							} else {
								s.Challenge = code
								RedirectSignUpVerification(w, r)
								return
							}
						}
					}
				}
			}
			RedirectSignUp(w, r)
			return
		}
	}
}

func SignUpVerificationHandler(sessions SessionStore, users conveygo.UserStore, payments PaymentProcessor, welcomer EmailWelcomer, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path, r.Header)
		cookie, err := GetSignInSessionCookie(r)
		if err == nil && sessions.IsValidSignInSession(cookie.Value) {
			RedirectAccount(w, r)
			return
		}
		session := ""
		cookie, err = GetSignUpSessionCookie(r)
		if err == nil {
			session = cookie.Value
		}
		s := sessions.GetSignUpSession(session)
		if s == nil {
			RedirectSignUp(w, r)
			return
		}
		switch r.Method {
		case "GET":
			// Show sign-up-verification page
			data := struct {
				Error string
			}{
				Error: s.Error,
			}
			if err := template.Execute(w, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			}
			return
		case "POST":
			// Try sign user up
			s.Error = ""
			s.Verification = r.FormValue("verification")
			log.Println(r.Form) // TODO(v1) remove
			if s.Verification == "" || s.Verification != s.Challenge {
				s.Error = ERROR_INCORRECT_EMAIL_VERIFICATION
			} else {
				// Generate private key
				key, err := rsa.GenerateKey(rand.Reader, 4096)
				if err != nil {
					log.Println(err)
					s.Error = err.Error()
				} else {
					// Register alias
					if err := users.RegisterAlias(s.Alias, []byte(s.Password), key); err != nil {
						log.Println(err)
						s.Error = err.Error()
					} else {
						// Add key to keystore
						if err := users.AddKey(s.Alias, []byte(s.Password), key); err != nil {
							log.Println(err)
							s.Error = err.Error()
						} else {
							// Register Customer with Payment Processor
							customerId, err := payments.RegisterCustomer(s.Name, s.Email, s.Alias)
							if err != nil {
								log.Println(err)
								s.Error = err.Error()
							} else {
								// Register Customer with BC
								if err := users.RegisterCustomer(s.Alias, key, customerId); err != nil {
									log.Println(err)
									s.Error = err.Error()
								} else {
									if welcomer != nil {
										if err := welcomer.WelcomeEmail(s.Alias, s.Email); err != nil {
											log.Println(err)
										}
									}
									// TODO(v1) Mine transaction into BC which moves welcome tokens from server to customer
									/*
									   transaction := &conveygo.Transaction{
									       Sender:   merchant,
									       Receiver: customer,
									       Amount:   uint64(q),
									   }
									   log.Println("Transaction", transaction)
									   if err := s.Node.MineProto(transactions, bcgo.THRESHOLD_G, s.Listener, nil, nil, transaction); err != nil {
									       log.Println(err)
									       return
									   }
									*/
									// TODO(v3) Subscribe email to weely digest
									// Success!
									RedirectSignedUp(w, r)
									return
								}
							}
						}
					}
				}
			}
			RedirectSignUpVerification(w, r)
			return
		}
	}
}
