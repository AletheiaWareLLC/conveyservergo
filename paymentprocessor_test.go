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
	"github.com/AletheiaWareLLC/conveyservergo"
	"testing"
)

func makeMockPaymentProcessor(t *testing.T) main.PaymentProcessor {
	return &MockPaymentProcessor{
		PublishableKey: "foo",
		ClientSecret:   "bar",
	}
}

type MockPaymentProcessor struct {
	PublishableKey, ClientSecret string
}

func (m *MockPaymentProcessor) GetPublishableKey() string {
	return m.PublishableKey
}

func (m *MockPaymentProcessor) NewSetupIntent() (string, error) {
	return "", nil
}

func (m *MockPaymentProcessor) RegisterCustomer(name, email, alias string) (string, error) {
	return "", nil
}

func (m *MockPaymentProcessor) AddPaymentMethod(customerId, paymentMethodId string) (string, error) {
	return "", nil
}

func (m *MockPaymentProcessor) GetPaymentMethods(customerId string) ([]*main.PaymentMethod, error) {
	return nil, nil
}

func (m *MockPaymentProcessor) NewPaymentIntent(customerId, paymentMethodId, alias, description string, quantity, amount int64) (string, error) {
	return "", nil
}

func (m *MockPaymentProcessor) GetProducts(productIds []string) (map[string]*main.Product, error) {
	return nil, nil
}
