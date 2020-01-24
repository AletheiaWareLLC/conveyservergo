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
	"github.com/AletheiaWareLLC/conveygo"
	"html/template"
	"log"
	"net/http"
)

func AccountHandler(sessions SessionStore, ledger *conveygo.Ledger, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path, r.Header)
		// If not signed in, redirect to sign in page
		cookie, err := GetSignInSessionCookie(r)
		if err == nil {
			session := sessions.GetSignInSession(cookie.Value)
			if session != nil {
				id, err := sessions.RefreshSignInSession(session)
				if err == nil {
					http.SetCookie(w, CreateSignInSessionCookie(id, sessions.GetSignInSessionTimeout()))
				}
				switch r.Method {
				case "GET":
					// Show account page
					data := struct {
						Alias   string
						Balance int64
					}{
						Alias:   session.Alias,
						Balance: ledger.GetBalance(session.Alias),
					}

					if err := template.Execute(w, data); err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
					return
				case "POST":
					// TODO
					return
				default:
					log.Println("Unsupported method", r.Method)
				}
			}
		}
		RedirectSignIn(w, r)
	}
}

type AddPaymentMethodTemplate struct {
	Error          string
	PublishableKey string
	ClientSecret   string
}

func AddPaymentMethodHandler(sessions SessionStore, users conveygo.UserStore, payments PaymentProcessor, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path, r.Header)
		cookie, err := GetSignInSessionCookie(r)
		if err == nil {
			session := sessions.GetSignInSession(cookie.Value)
			if session != nil {
				id, err := sessions.RefreshSignInSession(session)
				if err == nil {
					http.SetCookie(w, CreateSignInSessionCookie(id, sessions.GetSignInSessionTimeout()))
				}
				if session.AddPaymentMethod == nil {
					session.AddPaymentMethod = &AddPaymentMethodSession{}
				}
				s := session.AddPaymentMethod
				registration, err := users.GetRegistration(session.Alias)
				if registration == nil || err != nil {
					log.Println(registration)
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				switch r.Method {
				case "GET":
					data := &AddPaymentMethodTemplate{
						Error:          s.Error,
						PublishableKey: payments.GetPublishableKey(),
					}
					secret, err := payments.NewSetupIntent()
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
						return
					} else {
						data.ClientSecret = secret
					}
					if err := template.Execute(w, data); err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
					return
				case "POST":
					s.Error = ""
					s.PaymentMethod = r.FormValue("payment")
					// Attach Payment Method to Existing Customer
					if _, err := payments.AddPaymentMethod(registration.CustomerId, s.PaymentMethod); err != nil {
						log.Println(err)
						s.Error = err.Error()
						RedirectAddPaymentMethod(w, r)
					} else {
						// Success!
						session.AddPaymentMethod = nil
						RedirectAccount(w, r)
					}
					return
				default:
					log.Println("Unsupported method", r.Method)
				}
			}
		}
		RedirectSignIn(w, r)
	}
}
