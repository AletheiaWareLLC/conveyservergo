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
	"errors"
	"fmt"
	"github.com/AletheiaWareLLC/conveygo"
	"html/template"
	"regexp"
)

const ERROR_UNRECOGNIZED_MEDIA_TYPE = "Unrecognized Media Type: %s"

var (
	newlines = regexp.MustCompile(`\r?\n\r?\n`)
	anchors  = regexp.MustCompile(`\b(file|ftp|https?):\/\/\S+[\/\w]`)
)

func ContentToHTML(message *conveygo.Message) (template.HTML, error) {
	switch message.GetType() {
	case conveygo.MediaType_TEXT_PLAIN:
		safe := template.HTMLEscapeString(string(message.GetContent()))
		safe = anchors.ReplaceAllString(safe, `<a href="$0">$0</a>`)
		safe = newlines.ReplaceAllString(safe, `</p><p>`)
		return template.HTML(`<p>` + safe + `</p>`), nil
	default:
		return "", errors.New(fmt.Sprintf(ERROR_UNRECOGNIZED_MEDIA_TYPE, message.GetType()))
	}
}
