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

type PaymentProcessor interface {
	GetPublishableKey() string
	NewSetupIntent() (string, error)
	RegisterCustomer(name, email, alias string) (string, error)
	AddPaymentMethod(customerId, paymentMethodId string) (string, error)
	GetPaymentMethods(customerId string) ([]*PaymentMethod, error)
	NewPaymentIntent(customerId, paymentMethodId, alias, description string, quantity, amount int64) (string, error)
	GetProducts(productIds []string) (map[string]*Product, error)
}

type PaymentMethod struct {
	ID             string
	BillingDetails *BillingDetails
	Card           *PaymentMethodCard
}

type BillingDetails struct {
	Address *Address
	Email   string
	Name    string
	Phone   string
}

type Address struct {
	City       string
	Country    string
	Line1      string
	Line2      string
	PostalCode string
	State      string
}

type PaymentMethodCard struct {
	Brand    string
	Country  string
	ExpMonth uint64
	ExpYear  uint64
	Funding  string
	Last4    string
}

type Product struct {
	ID       string
	Quantity uint64
	Price    uint64
}
