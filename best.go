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
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

func BestHandler(sessions SessionStore, messages conveygo.MessageStore, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
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
			now := time.Now()
			var since uint64
			period := r.FormValue("period")
			switch period {
			case "all":
				since = 0
			case "year":
				since = uint64(now.Truncate(8766 * time.Hour).UnixNano())
			default:
				period = "day"
				fallthrough
			case "day":
				since = uint64(now.Truncate(24 * time.Hour).UnixNano())
			}
			limit := uint(8)
			l := r.FormValue("limit")
			if l != "" {
				if i, err := strconv.Atoi(l); err != nil {
					log.Println(err)
				} else {
					limit = uint(i)
				}
			}

			conversations, err := messages.GetAllConversations(since)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			var listings []*ConversationTemplate
			for _, c := range conversations {
				listing := &ConversationTemplate{
					ConversationHash: base64.RawURLEncoding.EncodeToString(c.Hash),
					Topic:            c.Topic,
					Timestamp:        bcgo.TimestampToString(c.Timestamp),
					Cost:             c.Cost,
					Author:           c.Author,
				}
				cost, reward, err := messages.GetYield(c.Hash)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				listing.Cost += cost
				listing.Reward += reward
				listing.Yield = int64(listing.Reward) - int64(listing.Cost)
				listings = append(listings, listing)
			}

			sort.Slice(listings, func(i, j int) bool {
				return listings[i].Yield > listings[j].Yield
			})

			if uint(len(listings)) > limit {
				listings = listings[:limit]
			}

			data := struct {
				Listings []*ConversationTemplate
				Period   string
				Limit    uint
			}{
				Listings: listings,
				Period:   period,
				Limit:    limit * 2,
			}

			if err := template.Execute(w, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			return
		default:
			log.Println("Unsupported method", r.Method)
			return
		}
	}
}
