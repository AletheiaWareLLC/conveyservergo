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
	"github.com/AletheiaWareLLC/conveygo"
	"html/template"
	"log"
	"net/http"
	"sort"
)

type LedgerTemplate struct {
	Entries []*LedgerEntryTemplate
	Sort    string
}

type LedgerEntryTemplate struct {
	Alias   string
	Minted  uint64 // Miner in Periodic Validation
	Burned  uint64 // Author in Conversation Chain, or First Message in Message Chain
	Bought  uint64 // Recipient in Transaction Chain
	Sold    uint64 // Sender in Transaction Chain
	Earned  uint64 // Got replied to in Message Chain
	Spent   uint64 // Creator in Conversation or Message Chain
	Balance int64
}

func LedgerHandler(ledger *conveygo.Ledger, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path, r.Header)
		switch r.Method {
		case "GET":
			s := r.FormValue("sort")
			var sorter func(*LedgerEntryTemplate, *LedgerEntryTemplate) bool
			switch s {
			case "alias":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Alias > j.Alias
				}
			case "minted":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Minted > j.Minted
				}
			case "burned":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Burned > j.Burned
				}
			case "bought":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Bought > j.Bought
				}
			case "sold":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Sold > j.Sold
				}
			case "earned":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Earned > j.Earned
				}
			case "spent":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Spent > j.Spent
				}
			default:
				s = "balance"
				fallthrough
			case "balance":
				sorter = func(i, j *LedgerEntryTemplate) bool {
					return i.Balance > j.Balance
				}
			}

			var entries []*LedgerEntryTemplate
			for alias := range ledger.Aliases {
				entries = append(entries, &LedgerEntryTemplate{
					Alias:   alias,
					Minted:  ledger.Minted[alias],
					Burned:  ledger.Burned[alias],
					Bought:  ledger.Bought[alias],
					Sold:    ledger.Sold[alias],
					Earned:  ledger.Earned[alias],
					Spent:   ledger.Spent[alias],
					Balance: ledger.GetBalance(alias),
				})
			}

			sort.Slice(entries, func(i, j int) bool {
				return sorter(entries[i], entries[j])
			})

			data := &LedgerTemplate{
				Entries: entries,
				Sort:    s,
			}

			if err := template.Execute(w, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			}
		default:
			log.Println("Unsupported method", r.Method)
		}
	}
}
