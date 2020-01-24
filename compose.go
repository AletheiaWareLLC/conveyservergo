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
	"encoding/base64"
	"errors"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/conveygo"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type ComposeTemplate struct {
	ConversationHash string
	MessageHash      string
	Timestamp        string
	Cost             uint64
	Reward           uint64
	Author           string
	Content          template.HTML
}

func CreateComposeTemplate(messages conveygo.MessageStore, conversation, message string) (*ComposeTemplate, error) {
	data := &ComposeTemplate{
		ConversationHash: conversation,
		MessageHash:      message,
	}

	if conversation != "" && message != "" {
		conversationHash, err := base64.RawURLEncoding.DecodeString(conversation)
		if err != nil {
			return nil, err
		}
		messageHash, err := base64.RawURLEncoding.DecodeString(message)
		if err != nil {
			return nil, err
		}
		if err := messages.GetMessage(conversationHash, messageHash, func(hash []byte, timestamp uint64, author string, cost uint64, message *conveygo.Message) error {
			content, err := ContentToHTML(message)
			if err != nil {
				return err
			} else {
				data.Timestamp = bcgo.TimestampToString(timestamp)
				data.Author = author
				data.Cost = cost
				data.Content = content
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func CreateDraftContributionSession(alias string, key *rsa.PrivateKey, conversation, message, topic, content string) (*DraftContributionSession, error) {
	timestamp := bcgo.Timestamp()
	draft := &DraftContributionSession{}

	if conversation == "" {
		if topic == "" {
			topic = bcgo.TimestampToString(timestamp)
		}
		draft.Conversation = &conveygo.Conversation{
			Topic: topic,
		}
		conversationHash, conversationRecord, err := conveygo.ProtoToRecord(alias, key, timestamp, draft.Conversation)
		if err != nil {
			return nil, err
		}
		draft.ConversationHash = conversationHash
		draft.ConversationRecord = conversationRecord
		draft.ConversationCost = conveygo.Cost(conversationRecord)

		draft.Message = &conveygo.Message{
			Content: []byte(content),
			Type:    conveygo.MediaType_TEXT_PLAIN,
		}
		messageHash, messageRecord, err := conveygo.ProtoToRecord(alias, key, timestamp, draft.Message)
		if err != nil {
			return nil, err
		}
		draft.MessageHash = messageHash
		draft.MessageRecord = messageRecord
		draft.MessageCost = conveygo.Cost(messageRecord)
	} else if message != "" {
		conversationHash, err := base64.RawURLEncoding.DecodeString(conversation)
		if err != nil {
			return nil, err
		}
		draft.ConversationHash = conversationHash

		previousHash, err := base64.RawURLEncoding.DecodeString(message)
		if err != nil {
			return nil, err
		}
		// TODO(v2) error if previous hash doesn't exist
		draft.Message = &conveygo.Message{
			Previous: previousHash,
			Content:  []byte(content),
			Type:     conveygo.MediaType_TEXT_PLAIN,
		}

		messageHash, messageRecord, err := conveygo.ProtoToRecord(alias, key, timestamp, draft.Message)
		if err != nil {
			return nil, err
		}
		draft.MessageHash = messageHash
		draft.MessageRecord = messageRecord
		draft.MessageCost = conveygo.Cost(messageRecord)
	} else {
		return nil, errors.New("Missing Message Hash")
	}
	return draft, nil
}

func ComposeHandler(sessions SessionStore, messages conveygo.MessageStore, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
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
				conversation := strings.TrimSpace(r.FormValue("conversation")) // Record hash of conversation being commented on, "" if new conversation
				message := strings.TrimSpace(r.FormValue("message"))           // Record hash of comment being replied to, "" if new conversation
				switch r.Method {
				case "GET":
					data, err := CreateComposeTemplate(messages, conversation, message)
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
						return
					}

					if err := template.Execute(w, data); err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
					return
				case "POST":
					topic := strings.TrimSpace(r.FormValue("topic")) // Topic of conversation being published, "" if reply
					content := strings.TrimSpace(r.FormValue("content"))

					draft, err := CreateDraftContributionSession(session.Alias, session.Key, conversation, message, topic, content)
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
						return
					}

					session.DraftContribution = draft
					RedirectPreview(w, r)
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
