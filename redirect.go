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
	"net/http"
)

func RedirectAccount(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/account", http.StatusFound)
}

func RedirectAddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	// TODO(v2) add a redirect parameter to send users to after they successfully add a payment method
	http.Redirect(w, r, "/add-payment-method.html", http.StatusFound)
}

func RedirectCompose(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/compose", http.StatusFound)
}

func RedirectConversation(w http.ResponseWriter, r *http.Request, conversation string) {
	http.Redirect(w, r, "/conversation?hash="+conversation, http.StatusFound)
}

func RedirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}

func RedirectPreview(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/preview", http.StatusFound)
}

func RedirectSignIn(w http.ResponseWriter, r *http.Request) {
	// TODO(v2) add a redirect parameter to send users to after they successfully sign in
	http.Redirect(w, r, "/sign-in", http.StatusFound)
}

func RedirectSignUp(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/sign-up", http.StatusFound)
}

func RedirectSignUpVerification(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/sign-up-verification", http.StatusFound)
}

func RedirectTokenPurchase(w http.ResponseWriter, r *http.Request) {
	// TODO(v2) add a redirect parameter to send users to after they successfully purchase tokens
	http.Redirect(w, r, "/token-purchase", http.StatusFound)
}

func RedirectTokenTransfer(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/token-transfer", http.StatusFound)
}

func RedirectPurchased(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/purchased.html", http.StatusFound)
}

func RedirectRegistered(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/registered.html", http.StatusFound)
}

func RedirectSignedUp(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/signed-up.html", http.StatusFound)
}

func RedirectSubscribed(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/subscribed.html", http.StatusFound)
}

func RedirectTransfered(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/transfered.html", http.StatusFound)
}
