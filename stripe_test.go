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
	"github.com/AletheiaWareLLC/testinggo"
	"net/http"
	"testing"
)

func TestStripeHandler(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		// TODO(v1)
	})
	t.Run("Single", func(t *testing.T) {
		// TODO(v1)
	})
	t.Run("Multiple", func(t *testing.T) {
		// TODO(v1)
	})
}

func makeGetStripeRequest(t *testing.T) *http.Request {
	request, err := http.NewRequest(http.MethodGet, "/token", nil)
	testinggo.AssertNoError(t, err)
	return request
}
