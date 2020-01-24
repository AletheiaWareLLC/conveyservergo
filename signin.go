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

type SignInTemplate struct {
	Error string
}

func SignInHandler(sessions SessionStore, users conveygo.UserStore, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path, r.Header)
		// If not signed in, show sign-in page
		cookie, err := GetSignInSessionCookie(r)
		if err != nil || !sessions.IsValidSignInSession(cookie.Value) {
			switch r.Method {
			case "GET":
				// Show sign-in page
				data := SignInTemplate{}
				if err := template.Execute(w, data); err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				}
				return
			case "POST":
				// Try sign user in
				alias := r.FormValue("alias")
				password := r.FormValue("password")

				key, err := users.GetKey(alias, []byte(password))
				if err != nil {
					// TODO(v2) Give better error message and status code, consider redirecting to sign-in page
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				id, err := sessions.CreateSignInSession(alias, key)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				http.SetCookie(w, CreateSignInSessionCookie(id, sessions.GetSignInSessionTimeout()))
			}
		}
		RedirectAccount(w, r)
	}
}
