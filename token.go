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
	"fmt"
	"github.com/AletheiaWareLLC/aliasgo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/golang/protobuf/proto"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
)

/* Token Bundles:
- 25 tokens for 50c ($0.02 each)
- 100 tokens for $1 ($0.01 each)
- 1250 tokens for $10 ($0.008 each)
- 20000 tokens for $100 ($0.005 each)
- 1000000 tokens for $1000 ($0.001 each)
*/

const (
	ERROR_INVALID_TOKEN_QUANTITY      = "Invalid token quantity: %d"
	ERROR_NO_SUCH_ALIAS               = "No such alias: %s"
	ERROR_NO_SUCH_PRODUCT             = "No such product: %s"
	ERROR_NOT_ENOUGH_TOKENS_AVAILABLE = "Not enough tokens available: %d requested, %d available"
)

type TokenPurchaseTemplate struct {
	Error         string
	TokenBundle   []*TokenBundle
	PaymentMethod []*PaymentMethod
}

type TokenBundle struct {
	ID        string
	Quantity  uint64
	Price     string
	UnitPrice string
}

func TokenPurchaseHandler(sessions SessionStore, users conveygo.UserStore, payments PaymentProcessor, ledger *conveygo.Ledger, node *bcgo.Node, template *template.Template, productIds []string) func(w http.ResponseWriter, r *http.Request) {
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
				registration, err := users.GetRegistration(session.Alias)
				if err != nil {
					// TODO(v2) if user exists but not registered as customer offer registration flow
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				available := ledger.GetBalance(node.Alias)
				if session.TokenPurchase == nil {
					session.TokenPurchase = &TokenPurchaseSession{}
				}
				s := session.TokenPurchase
				switch r.Method {
				case "GET":
					s.Product, err = payments.GetProducts(productIds)
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
						return
					}
					data := &TokenPurchaseTemplate{
						Error: s.Error,
					}
					for _, p := range s.Product {
						if p.Quantity <= uint64(available) {
							price := float64(p.Price)
							unit := price / float64(p.Quantity)
							data.TokenBundle = append(data.TokenBundle, &TokenBundle{
								ID:        p.ID,
								Quantity:  p.Quantity,
								Price:     fmt.Sprintf("$%.2f", price/100),
								UnitPrice: fmt.Sprintf("$%s", strconv.FormatFloat(unit/100, 'f', -1, 64)),
							})
						}
					}
					sort.Slice(data.TokenBundle, func(i, j int) bool {
						return data.TokenBundle[i].Quantity < data.TokenBundle[j].Quantity
					})
					if registration != nil {
						data.PaymentMethod, err = payments.GetPaymentMethods(registration.CustomerId)
						if err != nil {
							log.Println(err)
							http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
							return
						}
					}
					if err := template.Execute(w, data); err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
					return
				case "POST":
					s.Error = ""
					productId := r.FormValue("product")
					paymentMethodId := r.FormValue("payment-method")

					product, ok := s.Product[productId]
					if !ok {
						s.Error = fmt.Sprintf(ERROR_NO_SUCH_PRODUCT, productId)
						RedirectTokenPurchase(w, r)
						return
					}
					if product.Quantity > uint64(available) {
						s.Error = fmt.Sprintf(ERROR_NOT_ENOUGH_TOKENS_AVAILABLE, product.Quantity, available)
						RedirectTokenPurchase(w, r)
						return
					}

					_, err := payments.NewPaymentIntent(registration.CustomerId, paymentMethodId, session.Alias, fmt.Sprintf("%d Convey Tokens", product.Quantity), int64(product.Quantity), int64(product.Price))
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
						return
					}
					RedirectPurchased(w, r)
					return
				default:
					log.Println("Unsupported method", r.Method)
				}
			}
		}
		RedirectSignIn(w, r)
	}
}

func TokenSubscriptionHandler(sessions SessionStore, users conveygo.UserStore, payments PaymentProcessor, node *bcgo.Node, template *template.Template, productId, planId string) func(w http.ResponseWriter, r *http.Request) {
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
				switch r.Method {
				case "GET":
					// TODO(v3)
					return
				case "POST":
					// TODO(v3)
					RedirectSubscribed(w, r)
					return
				default:
					log.Println("Unsupported method", r.Method)
				}
			}
		}
		RedirectSignIn(w, r)
	}
}

type TokenTransferTemplate struct {
	Error     string
	Available int64
}

func TokenTransferHandler(sessions SessionStore, users conveygo.UserStore, ledger *conveygo.Ledger, aliases, transactions *bcgo.Channel, node *bcgo.Node, listener bcgo.MiningListener, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
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
				available := ledger.GetBalance(session.Alias)
				if session.TokenTransfer == nil {
					session.TokenTransfer = &TokenTransferSession{}
				}
				s := session.TokenTransfer
				switch r.Method {
				case "GET":
					data := &TokenTransferTemplate{
						Error:     s.Error,
						Available: available,
					}
					if err := template.Execute(w, data); err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					}
					return
				case "POST":
					s.Error = ""
					quantity := r.FormValue("quantity")
					recipient := r.FormValue("recipient")

					q, err := strconv.Atoi(quantity)
					if err != nil {
						s.Error = err.Error()
					} else if q <= 0 {
						s.Error = fmt.Sprintf(ERROR_INVALID_TOKEN_QUANTITY, q)
					} else if int64(q) > available {
						s.Error = fmt.Sprintf(ERROR_NOT_ENOUGH_TOKENS_AVAILABLE, q, available)
					} else {
						_, err := aliasgo.GetPublicKey(aliases, node.Cache, node.Network, recipient)
						if err != nil {
							s.Error = fmt.Sprintf(ERROR_NO_SUCH_ALIAS, recipient)
						} else {
							if err := MineTransaction(node, listener, transactions, session.Alias, session.Key, recipient, uint64(q)); err != nil {
								s.Error = err.Error()
							} else {
								RedirectTransfered(w, r)
								return
							}
						}
					}
					RedirectTokenTransfer(w, r)
					return
				default:
					log.Println("Unsupported method", r.Method)
				}
			}
		}
		RedirectSignIn(w, r)
	}
}

func MineTransaction(node *bcgo.Node, listener bcgo.MiningListener, transactions *bcgo.Channel, senderAlias string, senderKey *rsa.PrivateKey, recipient string, amount uint64) error {
	transaction := &conveygo.Transaction{
		Sender:   senderAlias,
		Receiver: recipient,
		Amount:   amount,
	}
	log.Println("Transaction", transaction)

	data, err := proto.Marshal(transaction)
	if err != nil {
		return err
	}

	_, record, err := bcgo.CreateRecord(bcgo.Timestamp(), senderAlias, senderKey, nil, nil, data)
	if err != nil {
		return err
	}
	log.Println("Record", record)

	if _, err := bcgo.WriteRecord(transactions.Name, node.Cache, record); err != nil {
		return err
	}

	if _, _, err := node.Mine(transactions, bcgo.THRESHOLD_G, listener); err != nil {
		return err
	}

	if err := transactions.Push(node.Cache, node.Network); err != nil {
		return err
	}
	return nil
}
