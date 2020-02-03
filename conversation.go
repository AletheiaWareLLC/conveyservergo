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
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/conveygo/html"
	"html/template"
	"log"
	"net/http"
	"sort"
)

type ConversationTemplate struct {
	ConversationHash string
	MessageHash      string
	Topic            string
	Timestamp        string
	Cost             uint64
	Reward           uint64
	Yield            int64
	Author           string
	Content          template.HTML
	Replies          []*ReplyTemplate
}

type ReplyTemplate struct {
	ConversationHash string
	PreviousHash     string
	MessageHash      string
	Timestamp        string
	Cost             uint64
	Reward           uint64
	Yield            int64
	Author           string
	Content          template.HTML
	Replies          []*ReplyTemplate
}

func ConversationHandler(sessions SessionStore, messages conveygo.MessageStore, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
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
			}
		}
		switch r.Method {
		case "GET":
			h := r.FormValue("hash")
			if h == "" {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			hash, err := base64.RawURLEncoding.DecodeString(h)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			listing, err := messages.GetConversation(hash)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			data := &ConversationTemplate{
				ConversationHash: h,
				Topic:            listing.Topic,
				Timestamp:        bcgo.TimestampToString(listing.Timestamp),
				Cost:             listing.Cost,
				Author:           listing.Author,
			}

			// TODO(v3) limit message count and add "more" button
			// TODO(v5) if message is encrypted only show if user is signed in, and alias is granted access
			replies := make(map[string]*ReplyTemplate)
			if err := messages.GetMessage(hash, nil, func(hash []byte, timestamp uint64, author string, cost uint64, message *conveygo.Message) error {
				content, err := html.ContentToHTML(message)
				if err != nil {
					return err
				}
				key := base64.RawURLEncoding.EncodeToString(hash)
				if message.Previous == nil || len(message.Previous) == 0 {
					data.MessageHash = key
					data.Content = content
					data.Cost += cost
				} else {
					replies[key] = &ReplyTemplate{
						ConversationHash: h,
						PreviousHash:     base64.RawURLEncoding.EncodeToString(message.Previous),
						MessageHash:      key,
						Timestamp:        bcgo.TimestampToString(timestamp),
						Cost:             cost,
						Yield:            -int64(cost),
						Author:           author,
						Content:          content,
					}
				}
				return nil
			}); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			for _, v := range replies {
				// Create reply tree
				if v.PreviousHash == data.MessageHash {
					data.Replies = append(data.Replies, v)
				} else {
					previous := replies[v.PreviousHash]
					previous.Replies = append(previous.Replies, v)
				}
				// Calculate rewards
				reply := v
				reward := v.Cost
				for reply != nil && reward > 0 {
					half := reward / 2 // Integer division so half of 3 is 1
					if reply.PreviousHash == data.MessageHash {
						data.Reward += half // Half of cost becomes reward
						// Remaining tokens are burned
						break
					} else {
						previous := replies[reply.PreviousHash]
						previous.Reward += half // Half of cost becomes reward
						previous.Yield = int64(previous.Reward) - int64(previous.Cost)
						reply = previous
						reward -= half // Remaining tokens go up the hierarchy
					}
				}
			}

			data.Yield = int64(data.Reward) - int64(data.Cost)

			// Sort replies by timestamp or yield
			RecursiveSort(data.Replies)

			if err := template.Execute(w, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			return
		default:
			log.Println("Unsupported method", r.Method)
		}
	}
}

func RecursiveSort(replies []*ReplyTemplate) {
	sort.Slice(replies, func(i, j int) bool {
		if replies[i].Yield == replies[j].Yield {
			return replies[i].Timestamp > replies[j].Timestamp
		}
		return replies[i].Yield > replies[j].Yield
	})
	for _, r := range replies {
		RecursiveSort(r.Replies)
	}
}
