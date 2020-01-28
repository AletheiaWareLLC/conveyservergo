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

package main_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/conveyservergo"
	"github.com/AletheiaWareLLC/testinggo"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeConversationTemplate(t *testing.T) *template.Template {
	t.Helper()
	tmplt, err := template.New("").Parse(`{{ .ConversationHash }}{{ .Timestamp }}{{ .Topic }}{{ .Content }}{{ define "reply" }}{{ range .Replies }}{{ .Timestamp }}{{ .Content }}{{ if gt (len .Replies) 0 }}{{ template "reply" . }}{{ end }}{{ end }}{{ end }}{{ template "reply" . }}`)
	testinggo.AssertNoError(t, err)
	return tmplt
}

func TestConversationHandler(t *testing.T) {
	alias := "Alice"
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error("Could not generate key:", err)
	}
	t.Run("Exists", func(t *testing.T) {
		datastore := conveygo.NewMemoryStore()
		sessionstore := main.NewMemorySessionStore()
		conversationHash, conversationRecord, err := conveygo.ProtoToRecord(alias, key, 0, &conveygo.Conversation{
			Topic: "Test123",
		})
		testinggo.AssertNoError(t, err)
		messageHash, messageRecord, err := conveygo.ProtoToRecord(alias, key, 0, &conveygo.Message{
			Content: []byte("Foo"),
			Type:    conveygo.MediaType_TEXT_PLAIN,
		})
		testinggo.AssertNoError(t, err)
		testinggo.AssertNoError(t, datastore.NewConversation(conversationHash, conversationRecord, messageHash, messageRecord))

		conversationHashString := base64.RawURLEncoding.EncodeToString(conversationHash)

		request := makeGetConversationRequest(t, conversationHashString)
		response := httptest.NewRecorder()

		handler := main.ConversationHandler(sessionstore, datastore, makeConversationTemplate(t))
		handler(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
		}

		actual := response.Body.String()
		expected := conversationHashString + `1970-01-01 00:00:00Test123<p>Foo</p>`

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("Exists_Reply", func(t *testing.T) {
		datastore := conveygo.NewMemoryStore()
		sessionstore := main.NewMemorySessionStore()
		conversationHash, conversationRecord, err := conveygo.ProtoToRecord(alias, key, 0, &conveygo.Conversation{
			Topic: "Test123",
		})
		testinggo.AssertNoError(t, err)
		messageHash, messageRecord, err := conveygo.ProtoToRecord(alias, key, 0, &conveygo.Message{
			Content: []byte("Foo"),
			Type:    conveygo.MediaType_TEXT_PLAIN,
		})
		testinggo.AssertNoError(t, err)
		testinggo.AssertNoError(t, datastore.NewConversation(conversationHash, conversationRecord, messageHash, messageRecord))

		replyHash, replyRecord, err := conveygo.ProtoToRecord(alias, key, 1, &conveygo.Message{
			Previous: messageHash,
			Content:  []byte("Bar"),
			Type:     conveygo.MediaType_TEXT_PLAIN,
		})
		testinggo.AssertNoError(t, err)
		testinggo.AssertNoError(t, datastore.AddMessage(conversationHash, replyHash, replyRecord))

		conversationHashString := base64.RawURLEncoding.EncodeToString(conversationHash)

		request := makeGetConversationRequest(t, conversationHashString)
		response := httptest.NewRecorder()

		handler := main.ConversationHandler(sessionstore, datastore, makeConversationTemplate(t))
		handler(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
		}

		actual := response.Body.String()
		expected := conversationHashString + `1970-01-01 00:00:00Test123<p>Foo</p>1970-01-01 00:00:00<p>Bar</p>`

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("Exists_Replies", func(t *testing.T) {
		datastore := conveygo.NewMemoryStore()
		sessionstore := main.NewMemorySessionStore()
		conversationHash, conversationRecord, err := conveygo.ProtoToRecord(alias, key, 0, &conveygo.Conversation{
			Topic: "Test123",
		})
		testinggo.AssertNoError(t, err)
		messageHash, messageRecord, err := conveygo.ProtoToRecord(alias, key, 0, &conveygo.Message{
			Content: []byte("Foo"),
			Type:    conveygo.MediaType_TEXT_PLAIN,
		})
		testinggo.AssertNoError(t, err)
		testinggo.AssertNoError(t, datastore.NewConversation(conversationHash, conversationRecord, messageHash, messageRecord))

		reply1Hash, reply1Record, err := conveygo.ProtoToRecord(alias, key, 1, &conveygo.Message{
			Previous: messageHash,
			Content:  []byte("Bar"),
			Type:     conveygo.MediaType_TEXT_PLAIN,
		})
		testinggo.AssertNoError(t, err)
		testinggo.AssertNoError(t, datastore.AddMessage(conversationHash, reply1Hash, reply1Record))

		reply2Hash, reply2Record, err := conveygo.ProtoToRecord(alias, key, 2, &conveygo.Message{
			Previous: reply1Hash,
			Content:  []byte("Baz"),
			Type:     conveygo.MediaType_TEXT_PLAIN,
		})
		testinggo.AssertNoError(t, err)
		testinggo.AssertNoError(t, datastore.AddMessage(conversationHash, reply2Hash, reply2Record))

		conversationHashString := base64.RawURLEncoding.EncodeToString(conversationHash)

		request := makeGetConversationRequest(t, conversationHashString)
		response := httptest.NewRecorder()

		handler := main.ConversationHandler(sessionstore, datastore, makeConversationTemplate(t))
		handler(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusOK, response.Code)
		}

		actual := response.Body.String()
		expected := conversationHashString + `1970-01-01 00:00:00Test123<p>Foo</p>1970-01-01 00:00:00<p>Bar</p>1970-01-01 00:00:00<p>Baz</p>`

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
	t.Run("NotExists", func(t *testing.T) {
		datastore := conveygo.NewMemoryStore()
		sessionstore := main.NewMemorySessionStore()

		request := makeGetConversationRequest(t, "foobar123456")
		response := httptest.NewRecorder()

		handler := main.ConversationHandler(sessionstore, datastore, makeConversationTemplate(t))
		handler(response, request)

		if response.Code != http.StatusNotFound {
			t.Errorf("Wrong response code; expected '%d', got '%d'", http.StatusNotFound, response.Code)
		}

		actual := response.Body.String()
		expected := "Not Found\n"

		if actual != expected {
			t.Errorf("Wrong response; expected '%s', got '%s'", expected, actual)
		}
	})
}

func makeGetConversationRequest(t *testing.T, hash string) *http.Request {
	request, err := http.NewRequest(http.MethodGet, "/conversation?hash="+hash, nil)
	testinggo.AssertNoError(t, err)
	return request
}
