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
	"html/template"
	"log"
	"net/http"
)

type SignOutTemplate struct {
}

func SignOutHandler(sessions SessionStore, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path, r.Header)
		// If signed in, show sign-out page
		cookie, err := GetSignInSessionCookie(r)
		if err == nil && sessions.IsValidSignInSession(cookie.Value) {
			switch r.Method {
			case "GET":
				// Show sign-out page
				data := SignOutTemplate{}
				if err := template.Execute(w, data); err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				}
				return
			case "POST":
				// Try sign user out
				sessions.DeleteSignInSession(cookie.Value)
				RedirectHome(w, r)
				return
			}
		}
		RedirectSignIn(w, r)
	}
}
