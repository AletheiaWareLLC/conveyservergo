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
	"github.com/AletheiaWareLLC/conveygo/html"
	"html/template"
	"log"
	"net/http"
)

type PreviewTemplate struct {
	Topic   string
	Content template.HTML
	Balance int64
	Cost    uint64
}

func PreviewHandler(sessions SessionStore, messages conveygo.MessageStore, ledger *conveygo.Ledger, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
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
				draft := session.DraftContribution
				if draft == nil {
					// No draft to preview, redirect to compose page
					RedirectCompose(w, r)
					return
				}
				switch r.Method {
				case "GET":
					data := &PreviewTemplate{
						Balance: ledger.GetBalance(session.Alias),
						Cost:    draft.MessageCost,
					}

					content, err := html.ContentToHTML(draft.Message)
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
					data.Content = content

					if draft.Conversation != nil {
						data.Topic = draft.Conversation.Topic
						data.Cost += draft.ConversationCost
					}

					if err := template.Execute(w, data); err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
					return
				default:
					log.Println("Unsupported method", r.Method)
					return
				}
			}
		}
		RedirectSignIn(w, r)
	}
}
