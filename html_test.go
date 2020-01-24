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
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/conveyservergo"
	"github.com/AletheiaWareLLC/testinggo"
	"testing"
)

func testTextPlain(t *testing.T, expected, content string) {
	html, err := main.ContentToHTML(&conveygo.Message{
		Content: []byte(content),
		Type:    conveygo.MediaType_TEXT_PLAIN,
	})
	testinggo.AssertNoError(t, err)
	actual := string(html)
	if actual != expected {
		t.Errorf("Wrong HTML; expected '%s', got '%s'", expected, actual)
	}
}

func TestContentToHTML(t *testing.T) {
	t.Run("Text_Plain", func(t *testing.T) {
		expected := `<p>FooBar</p>`
		testTextPlain(t, expected, "FooBar")
	})
	t.Run("Text_Plain_Paragraph", func(t *testing.T) {
		expected := `<p>Foo</p><p>Bar</p>`
		testTextPlain(t, expected, "Foo\n\nBar")
	})
	t.Run("Text_Plain_URL", func(t *testing.T) {
		expected := `<p><a href="https://example.com">https://example.com</a></p>`
		testTextPlain(t, expected, "https://example.com")
	})
	t.Run("Text_Plain_URL_User", func(t *testing.T) {
		expected := `<p><a href="https://alice@example.com">https://alice@example.com</a></p>`
		testTextPlain(t, expected, "https://alice@example.com")
	})
	t.Run("Text_Plain_URL_Port", func(t *testing.T) {
		expected := `<p><a href="https://example.com:8080">https://example.com:8080</a></p>`
		testTextPlain(t, expected, "https://example.com:8080")
	})
	t.Run("Text_Plain_URL_Query", func(t *testing.T) {
		expected := `<p><a href="https://example.com/test?foo=bar">https://example.com/test?foo=bar</a></p>`
		testTextPlain(t, expected, "https://example.com/test?foo=bar")
	})
	t.Run("Text_Plain_URL_Fragment", func(t *testing.T) {
		expected := `<p><a href="https://example.com/test#foobar">https://example.com/test#foobar</a></p>`
		testTextPlain(t, expected, "https://example.com/test#foobar")
	})
	t.Run("Text_Plain_URL_In_Text", func(t *testing.T) {
		expected := `<p>Visit <a href="https://example.com">https://example.com</a> for more.</p>`
		testTextPlain(t, expected, "Visit https://example.com for more.")
	})
}
