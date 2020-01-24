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
	"encoding/base64"
	"github.com/AletheiaWareLLC/conveygo"
	"html/template"
	"log"
	"net/http"
)

func PublishHandler(sessions SessionStore, messages conveygo.MessageStore, ledger *conveygo.Ledger, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
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
				case "POST":
					// Ensure user has sufficient balance to post
					balance := ledger.GetBalance(session.Alias)
					cost := draft.MessageCost
					if draft.Conversation != nil {
						cost += draft.ConversationCost
					}
					if balance < 0 || uint64(balance) < cost {
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
						return
					}

					var err error
					if draft.Conversation != nil {
						// Start new Conversation
						err = messages.NewConversation(draft.ConversationHash, draft.ConversationRecord, draft.MessageHash, draft.MessageRecord)
					} else {
						// Add Message to existing Conversation
						err = messages.AddMessage(draft.ConversationHash, draft.MessageHash, draft.MessageRecord)
					}
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					} else {
						session.DraftContribution = nil
						RedirectConversation(w, r, base64.RawURLEncoding.EncodeToString(draft.ConversationHash))
					}
					ledger.TriggerUpdate()
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
