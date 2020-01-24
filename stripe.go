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
	"github.com/AletheiaWareLLC/aliasgo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/financego"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/paymentmethod"
	"github.com/stripe/stripe-go/product"
	"github.com/stripe/stripe-go/setupintent"
	"log"
	"strconv"
)

const (
	META_ALIAS_MERCHANT  = "merchant_alias"
	META_ALIAS_CUSTOMER  = "customer_alias"
	META_QUANTITY_TOKENS = "token_quantity"
	META_BUNDLE_PRICE    = "bundle_price"
)

type StripePaymentProcessor struct {
	PublishableKey string
	Node           *bcgo.Node
	Listener       bcgo.MiningListener
}

func NewStripePaymentProcessor(secret, publishable string, node *bcgo.Node, listener bcgo.MiningListener) *StripePaymentProcessor {
	stripe.Key = secret
	return &StripePaymentProcessor{
		PublishableKey: publishable,
		Node:           node,
		Listener:       listener,
	}
}

func (s *StripePaymentProcessor) GetPublishableKey() string {
	return s.PublishableKey
}

func (s *StripePaymentProcessor) NewSetupIntent() (string, error) {
	intent, err := setupintent.New(&stripe.SetupIntentParams{})
	if err != nil {
		return "", err
	}
	return intent.ClientSecret, nil
}

func (s *StripePaymentProcessor) RegisterCustomer(name, email, alias string) (string, error) {
	params := &stripe.CustomerParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}
	params.AddMetadata(META_ALIAS_MERCHANT, s.Node.Alias)
	params.AddMetadata(META_ALIAS_CUSTOMER, alias)
	c, err := customer.New(params)
	if err != nil {
		return "", err
	}
	log.Println("Stripe Customer", c)
	return c.ID, nil
}

func (s *StripePaymentProcessor) AddPaymentMethod(customerId, paymentMethodId string) (string, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerId),
	}
	pm, err := paymentmethod.Attach(paymentMethodId, params)
	if err != nil {
		return "", err
	}
	log.Println("Stripe Payment Method", pm)
	return pm.ID, nil
}

func (s *StripePaymentProcessor) GetPaymentMethods(customerId string) ([]*PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerId),
		Type:     stripe.String("card"),
	}
	iterator := paymentmethod.List(params)
	var methods []*PaymentMethod
	for iterator.Next() {
		p := iterator.PaymentMethod()
		methods = append(methods, &PaymentMethod{
			Card: &PaymentMethodCard{
				Brand:    string(p.Card.Brand),
				Country:  p.Card.Country,
				ExpMonth: p.Card.ExpMonth,
				ExpYear:  p.Card.ExpYear,
				Funding:  string(p.Card.Funding),
				Last4:    p.Card.Last4,
			},
			ID: p.ID,
			BillingDetails: &BillingDetails{
				Address: &Address{
					City:       p.BillingDetails.Address.City,
					Country:    p.BillingDetails.Address.Country,
					Line1:      p.BillingDetails.Address.Line1,
					Line2:      p.BillingDetails.Address.Line2,
					PostalCode: p.BillingDetails.Address.PostalCode,
					State:      p.BillingDetails.Address.State,
				},
				Email: p.BillingDetails.Email,
				Name:  p.BillingDetails.Name,
				Phone: p.BillingDetails.Phone,
			},
		})
	}
	return methods, nil
}

func (s *StripePaymentProcessor) NewPaymentIntent(customerId, paymentMethodId, alias, description string, quantity, amount int64) (string, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      stripe.String(customerId),
		PaymentMethod: stripe.String(paymentMethodId),
		Description:   stripe.String(description),
		Confirm:       stripe.Bool(true),
		OffSession:    stripe.Bool(true),
	}
	params.AddMetadata(META_ALIAS_MERCHANT, s.Node.Alias)
	params.AddMetadata(META_ALIAS_CUSTOMER, alias)
	params.AddMetadata(META_QUANTITY_TOKENS, strconv.FormatInt(quantity, 10))

	c, err := paymentintent.New(params)
	if err != nil {
		return "", err
	}
	log.Println("Stripe Charge", c)
	return c.ID, nil
}

func (s *StripePaymentProcessor) GetProducts(productIds []string) (map[string]*Product, error) {
	var ids []*string
	for _, v := range productIds {
		ids = append(ids, stripe.String(v))
	}
	params := &stripe.ProductListParams{
		Active: stripe.Bool(true),
		IDs:    ids,
	}
	iterator := product.List(params)
	products := make(map[string]*Product)
	for iterator.Next() {
		p := iterator.Product()
		log.Println("ID", p.ID)
		log.Println("Active", p.Active)
		log.Println("Created", p.Created)
		log.Println("Updated", p.Updated)
		log.Println("Name", p.Name)
		log.Println("Description", p.Description)
		log.Println("Images", p.Images)
		log.Println("Metadata", p.Metadata)
		quantity, err := strconv.Atoi(p.Metadata[META_QUANTITY_TOKENS])
		if err != nil {
			return nil, err
		}
		price, err := strconv.Atoi(p.Metadata[META_BUNDLE_PRICE])
		if err != nil {
			return nil, err
		}
		products[p.ID] = &Product{
			ID:       p.ID,
			Quantity: uint64(quantity),
			Price:    uint64(price),
		}
	}
	return products, nil
}

func NewStripeEventHandler(aliases, charges, transactions *bcgo.Channel, node *bcgo.Node, listener bcgo.MiningListener) func(*stripe.Event) {
	return func(event *stripe.Event) {
		merchant := event.GetObjectValue("metadata", META_ALIAS_MERCHANT)
		log.Println("Merchant", merchant)
		if merchant == node.Alias {
			switch event.Type {
			case "account.updated":
				// TODO mine event into BC
			case "account.external_account.created":
				// TODO mine event into BC
			case "account.external_account.deleted":
				// TODO mine event into BC
			case "account.external_account.updated":
				// TODO mine event into BC
			case "balance.available":
				// TODO mine event into BC
			case "capability.updated":
				// TODO mine event into BC
			case "charge.captured":
				// TODO mine event into BC
			case "charge.expired":
				// TODO mine event into BC
			case "charge.failed":
				// TODO mine event into BC
			case "charge.pending":
				// TODO mine event into BC
			case "charge.refunded":
				// TODO mine event into BC
			case "charge.succeeded":
				// TODO mine event into BC
				customer := event.GetObjectValue("metadata", META_ALIAS_CUSTOMER)
				quantity := event.GetObjectValue("metadata", META_QUANTITY_TOKENS)
				amount := event.GetObjectValue("amount")
				chargeId := event.GetObjectValue("id")
				currency := event.GetObjectValue("currency")
				description := event.GetObjectValue("description")
				paymentId := event.GetObjectValue("payment_method")

				log.Println("Customer", customer)
				log.Println("Quantity", quantity)
				log.Println("Amount", amount)
				log.Println("ChargeId", chargeId)
				log.Println("Currency", currency)
				log.Println("Description", description)
				log.Println("PaymentId", paymentId)

				a, err := strconv.Atoi(amount)
				if err != nil {
					log.Println(err)
					return
				}
				q, err := strconv.Atoi(quantity)
				if err != nil {
					log.Println(err)
					return
				}

				publicKey, err := aliasgo.GetPublicKey(aliases, node.Cache, node.Network, customer)
				if err != nil {
					log.Println(err)
					return
				}

				charge := &financego.Charge{
					MerchantAlias: merchant,
					CustomerAlias: customer,
					Processor:     financego.PaymentProcessor_STRIPE,
					PaymentId:     paymentId,
					ChargeId:      chargeId,
					Amount:        int64(a),
					Currency:      currency,
					Description:   description,
				}
				log.Println("Charge", charge)
				if err := node.MineProto(charges, bcgo.THRESHOLD_G, listener, map[string]*rsa.PublicKey{
					customer: publicKey,
					merchant: &node.Key.PublicKey,
				}, nil, charge); err != nil {
					log.Println(err)
					return
				}

				transaction := &conveygo.Transaction{
					Sender:   merchant,
					Receiver: customer,
					Amount:   uint64(q),
				}
				log.Println("Transaction", transaction)
				if err := node.MineProto(transactions, bcgo.THRESHOLD_G, listener, nil, nil, transaction); err != nil {
					log.Println(err)
					return
				}
			case "charge.updated":
				// TODO mine event into BC
			case "charge.dispute.closed":
				// TODO mine event into BC
			case "charge.dispute.created":
				// TODO mine event into BC
			case "charge.dispute.funds_reinstated":
				// TODO mine event into BC
			case "charge.dispute.funds_withdrawn":
				// TODO mine event into BC
			case "charge.dispute.updated":
				// TODO mine event into BC
			case "charge.refund.updated":
				// TODO mine event into BC
			case "checkout.session.completed":
				// TODO mine event into BC
			case "coupon.created":
				// TODO mine event into BC
			case "coupon.deleted":
				// TODO mine event into BC
			case "coupon.updated":
				// TODO mine event into BC
			case "credit_note.created":
				// TODO mine event into BC
			case "credit_note.updated":
				// TODO mine event into BC
			case "credit_note.voided":
				// TODO mine event into BC
			case "customer.created":
				// TODO mine event into BC
				customer := event.GetObjectValue("metadata", META_ALIAS_CUSTOMER)
				log.Println("Customer", customer)
			case "customer.deleted":
				// TODO mine event into BC
			case "customer.updated":
				// TODO mine event into BC
			case "customer.bank_account.deleted":
				// TODO mine event into BC
			case "customer.discount.created":
				// TODO mine event into BC
			case "customer.discount.deleted":
				// TODO mine event into BC
			case "customer.discount.updated":
				// TODO mine event into BC
			case "customer.source.created":
				// TODO mine event into BC
			case "customer.source.deleted":
				// TODO mine event into BC
			case "customer.source.expiring":
				// TODO mine event into BC
			case "customer.source.updated":
				// TODO mine event into BC
			case "customer.subscription.created":
				// TODO mine event into BC
			case "customer.subscription.deleted":
				// TODO mine event into BC
			case "customer.subscription.trial_will_end":
				// TODO mine event into BC
			case "customer.subscription.updated":
				// TODO mine event into BC
			case "customer.tax_id.created":
				// TODO mine event into BC
			case "customer.tax_id.deleted":
				// TODO mine event into BC
			case "customer.tax_id.updated":
				// TODO mine event into BC
			case "file.created":
				// TODO mine event into BC
			case "invoice.created":
				// TODO mine event into BC
			case "invoice.deleted":
				// TODO mine event into BC
			case "invoice.finalized":
				// TODO mine event into BC
			case "invoice.marked_uncollectible":
				// TODO mine event into BC
			case "invoice.payment_action_required":
				// TODO mine event into BC
			case "invoice.payment_failed":
				// TODO mine event into BC
			case "invoice.payment_succeeded":
				// TODO mine event into BC
			case "invoice.sent":
				// TODO mine event into BC
			case "invoice.upcoming":
				// TODO mine event into BC
			case "invoice.updated":
				// TODO mine event into BC
			case "invoice.voided":
				// TODO mine event into BC
			case "invoiceitem.created":
				// TODO mine event into BC
			case "invoiceitem.deleted":
				// TODO mine event into BC
			case "invoiceitem.updated":
				// TODO mine event into BC
			case "issuing_authorization.created":
				// TODO mine event into BC
			case "issuing_authorization.request":
				// TODO mine event into BC
			case "issuing_authorization.updated":
				// TODO mine event into BC
			case "issuing_card.created":
				// TODO mine event into BC
			case "issuing_card.updated":
				// TODO mine event into BC
			case "issuing_cardholder.created":
				// TODO mine event into BC
			case "issuing_cardholder.updated":
				// TODO mine event into BC
			case "issuing_dispute.created":
				// TODO mine event into BC
			case "issuing_dispute.updated":
				// TODO mine event into BC
			case "issuing_settlement.created":
				// TODO mine event into BC
			case "issuing_settlement.updated":
				// TODO mine event into BC
			case "issuing_transaction.created":
				// TODO mine event into BC
			case "issuing_transaction.updated":
				// TODO mine event into BC
			case "order.created":
				// TODO mine event into BC
			case "order.payment_failed":
				// TODO mine event into BC
			case "order.payment_succeeded":
				// TODO mine event into BC
			case "order.updated":
				// TODO mine event into BC
			case "order_return.created":
				// TODO mine event into BC
			case "payment_intent.amount_capturable_updated":
				// TODO mine event into BC
			case "payment_intent.created":
				// TODO mine event into BC
			case "payment_intent.payment_failed":
				// TODO mine event into BC
			case "payment_intent.succeeded":
				// TODO mine event into BC
			case "payment_method.attached":
				// TODO mine event into BC
			case "payment_method.card_automatically_updated":
				// TODO mine event into BC
			case "payment_method.detached":
				// TODO mine event into BC
			case "payment_method.updated":
				// TODO mine event into BC
			case "payout.canceled":
				// TODO mine event into BC
			case "payout.created":
				// TODO mine event into BC
			case "payout.failed":
				// TODO mine event into BC
			case "payout.paid":
				// TODO mine event into BC
			case "payout.updated":
				// TODO mine event into BC
			case "person.created":
				// TODO mine event into BC
			case "person.deleted":
				// TODO mine event into BC
			case "person.updated":
				// TODO mine event into BC
			case "plan.created":
				// TODO mine event into BC
			case "plan.deleted":
				// TODO mine event into BC
			case "plan.updated":
				// TODO mine event into BC
			case "product.created":
				// TODO mine event into BC
			case "product.deleted":
				// TODO mine event into BC
			case "product.updated":
				// TODO mine event into BC
			case "radar.early_fraud_warning.created":
				// TODO mine event into BC
			case "radar.early_fraud_warning.updated":
				// TODO mine event into BC
			case "recipient.created":
				// TODO mine event into BC
			case "recipient.deleted":
				// TODO mine event into BC
			case "recipient.updated":
				// TODO mine event into BC
			case "reporting.report_run.failed":
				// TODO mine event into BC
			case "reporting.report_run.succeeded":
				// TODO mine event into BC
			case "reporting.report_type.updated":
				// TODO mine event into BC
			case "review.closed":
				// TODO mine event into BC
			case "review.opened":
				// TODO mine event into BC
			case "setup_intent.created":
				// TODO mine event into BC
			case "setup_intent.setup_failed":
				// TODO mine event into BC
			case "setup_intent.succeeded":
				// TODO mine event into BC
			case "sigma.scheduled_query_run.created":
				// TODO mine event into BC
			case "sku.created":
				// TODO mine event into BC
			case "sku.deleted":
				// TODO mine event into BC
			case "sku.updated":
				// TODO mine event into BC
			case "source.canceled":
				// TODO mine event into BC
			case "source.chargeable":
				// TODO mine event into BC
			case "source.failed":
				// TODO mine event into BC
			case "source.mandate_notification":
				// TODO mine event into BC
			case "source.refund_attributes_required":
				// TODO mine event into BC
			case "source.transaction.created":
				// TODO mine event into BC
			case "source.transaction.updated":
				// TODO mine event into BC
			case "tax_rate.created":
				// TODO mine event into BC
			case "tax_rate.updated":
				// TODO mine event into BC
			case "topup.canceled":
				// TODO mine event into BC
			case "topup.created":
				// TODO mine event into BC
			case "topup.failed":
				// TODO mine event into BC
			case "topup.reversed":
				// TODO mine event into BC
			case "topup.succeeded":
				// TODO mine event into BC
			case "transfer.created":
				// TODO mine event into BC
			case "transfer.failed":
				// TODO mine event into BC
			case "transfer.paid":
				// TODO mine event into BC
			case "transfer.reversed":
				// TODO mine event into BC
			case "transfer.updated":
				// TODO mine event into BC
			}
		}
	}
}
