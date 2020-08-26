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
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/AletheiaWareLLC/aliasgo"
	"github.com/AletheiaWareLLC/aliasservergo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/bcnetgo"
	"github.com/AletheiaWareLLC/conveygo"
	"github.com/AletheiaWareLLC/cryptogo"
	"github.com/AletheiaWareLLC/netgo"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type Server struct {
	Root     string
	Cert     string
	Cache    *bcgo.FileCache
	Network  *bcgo.TCPNetwork
	Listener bcgo.MiningListener
}

func (s *Server) Init() (*bcgo.Node, error) {
	// Add Convey hosts to peers
	for _, host := range conveygo.GetConveyHosts() {
		if err := bcgo.AddPeer(s.Root, host); err != nil {
			return nil, err
		}
	}

	// Add BC host to peers
	if err := bcgo.AddPeer(s.Root, bcgo.GetBCHost()); err != nil {
		return nil, err
	}

	// Create Node
	node, err := bcgo.GetNode(s.Root, s.Cache, s.Network)
	if err != nil {
		return nil, err
	}

	// Register Alias
	if err := aliasgo.Register(node, s.Listener); err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Server) LoadChannel(node *bcgo.Node, channel *bcgo.Channel) {
	if err := channel.Refresh(s.Cache, s.Network); err != nil {
		log.Println(err)
	}
	// Add channel to node
	node.AddChannel(channel)
}

func (s *Server) Start(node *bcgo.Node) error {
	// Create ledger
	ledger := conveygo.NewLedger(node)

	// Open channels
	aliases := aliasgo.OpenAliasChannel()
	hours := conveygo.OpenHourChannel()
	days := conveygo.OpenDayChannel()
	weeks := conveygo.OpenWeekChannel()
	years := conveygo.OpenYearChannel()
	decades := conveygo.OpenDecadeChannel()
	centuries := conveygo.OpenCenturyChannel()
	charges := conveygo.OpenChargeChannel()
	invoices := conveygo.OpenInvoiceChannel()
	registrations := conveygo.OpenRegistrationChannel()
	subscriptions := conveygo.OpenSubscriptionChannel()
	conversations := conveygo.OpenConversationChannel()
	transactions := conveygo.OpenTransactionChannel()

	// Add ledger triggers
	hours.AddTrigger(ledger.TriggerUpdate)
	days.AddTrigger(ledger.TriggerUpdate)  // TODO(v3) add digest trigger
	weeks.AddTrigger(ledger.TriggerUpdate) // TODO(v3) add digest trigger
	years.AddTrigger(ledger.TriggerUpdate) // TODO(v3) add digest trigger
	decades.AddTrigger(ledger.TriggerUpdate)
	centuries.AddTrigger(ledger.TriggerUpdate)
	conversations.AddTrigger(ledger.TriggerUpdate)
	transactions.AddTrigger(ledger.TriggerUpdate)

	// Create validators
	hourly := bcgo.GetHourlyValidator(hours)
	daily := bcgo.GetDailyValidator(days)
	weekly := bcgo.GetWeeklyValidator(weeks)
	yearly := bcgo.GetYearlyValidator(years)
	decennially := bcgo.GetDecenniallyValidator(decades)
	centennially := bcgo.GetCentenniallyValidator(centuries)

	for _, c := range []*bcgo.Channel{
		hours,
		days,
		weeks,
		years,
		decades,
		centuries,
		aliases,
		charges,
		invoices,
		registrations,
		subscriptions,
		conversations,
		transactions,
	} {
		// Add periodic validators
		c.AddValidator(hourly)
		c.AddValidator(daily)
		c.AddValidator(weekly)
		c.AddValidator(yearly)
		c.AddValidator(decennially)
		c.AddValidator(centennially)
		// Load channel
		s.LoadChannel(node, c)
	}

	channels := make(map[string]bool)
	// Mark all message channels listed in conversation channel
	if err := bcgo.Iterate(conversations.Name, conversations.Head, nil, s.Cache, s.Network, func(h []byte, b *bcgo.Block) error {
		for _, entry := range b.Entry {
			channels[conveygo.CONVEY_PREFIX_MESSAGE+base64.RawURLEncoding.EncodeToString(entry.RecordHash)] = true
		}
		return nil
	}); err != nil {
		return err
	}

	// Mark all channels listed in the periodic validation chains
	hourly.FillChannelSet(channels, s.Cache, s.Network)
	daily.FillChannelSet(channels, s.Cache, s.Network)
	weekly.FillChannelSet(channels, s.Cache, s.Network)
	yearly.FillChannelSet(channels, s.Cache, s.Network)
	decennially.FillChannelSet(channels, s.Cache, s.Network)
	centennially.FillChannelSet(channels, s.Cache, s.Network)

	// Unmark channels already open
	for k := range node.Channels {
		channels[k] = false
	}

	// Open all channels marked in map
	for c, b := range channels {
		if b && strings.HasPrefix(c, conveygo.CONVEY_PREFIX) {
			channel := bcgo.OpenPoWChannel(c, bcgo.THRESHOLD_G)
			if strings.HasPrefix(c, conveygo.CONVEY_PREFIX_MESSAGE) {
				channel.AddTrigger(ledger.TriggerUpdate)
			}
			s.LoadChannel(node, channel)
		}
	}

	go ledger.Start()
	ledger.TriggerUpdate()
	defer ledger.Stop()

	// Start Periodic Validation Chains
	go hourly.Start(node, bcgo.THRESHOLD_PERIOD_HOUR, s.Listener)
	defer hourly.Stop()
	go daily.Start(node, bcgo.THRESHOLD_PERIOD_DAY, s.Listener)
	defer daily.Stop()
	go weekly.Start(node, bcgo.THRESHOLD_PERIOD_WEEK, s.Listener)
	defer weekly.Stop()
	go yearly.Start(node, bcgo.THRESHOLD_PERIOD_YEAR, s.Listener)
	defer yearly.Stop()
	if bcgo.IsLive() {
		go decennially.Start(node, bcgo.THRESHOLD_PERIOD_DECADE, s.Listener)
		defer decennially.Stop()
		go centennially.Start(node, bcgo.THRESHOLD_PERIOD_CENTURY, s.Listener)
		defer centennially.Stop()
	}

	// Serve Block Requests
	go bcnetgo.BindTCP(bcgo.PORT_GET_BLOCK, bcnetgo.BlockPortTCPHandler(s.Cache, s.Network))
	// Serve Head Requests
	go bcnetgo.BindTCP(bcgo.PORT_GET_HEAD, bcnetgo.HeadPortTCPHandler(s.Cache, s.Network))
	// Serve Block Updates
	go bcnetgo.BindTCP(bcgo.PORT_BROADCAST, bcnetgo.BroadcastPortTCPHandler(s.Cache, s.Network, func(name string) (*bcgo.Channel, error) {
		channel, err := node.GetChannel(name)
		if err != nil {
			if strings.HasPrefix(name, conveygo.CONVEY_PREFIX) {
				channel = bcgo.OpenPoWChannel(name, bcgo.THRESHOLD_G)
				if strings.HasPrefix(name, conveygo.CONVEY_PREFIX_MESSAGE) {
					channel.AddTrigger(ledger.TriggerUpdate)
				}
				s.LoadChannel(node, channel)
			} else {
				return nil, err
			}
		}
		return channel, nil
	}))

	templates, err := template.ParseFiles(
		"html/template/account.go.html",
		// TODO(v2) "html/template/account-export.go.html",
		// TODO(v2) "html/template/account-import.go.html",
		"html/template/add-payment-method.go.html",
		"html/template/alias.go.html",
		"html/template/best.go.html",
		"html/template/block.go.html",
		"html/template/channel.go.html",
		"html/template/channel-list.go.html",
		"html/template/compose.go.html",
		"html/template/conversation.go.html",
		// TODO(v3) "html/template/digest.go.html",
		// TODO(v3) "html/template/email-digest.go.html",
		"html/template/email-verification.go.html",
		"html/template/email-welcome.go.html",
		"html/template/ledger.go.html",
		"html/template/listing.go.html",
		"html/template/message.go.html",
		"html/template/preview.go.html",
		"html/template/recent.go.html",
		"html/template/reply.go.html",
		"html/template/sign-in.go.html",
		"html/template/sign-out.go.html",
		"html/template/sign-up.go.html",
		"html/template/sign-up-verification.go.html",
		"html/template/token-purchase.go.html",
		// TODO(v3) "html/template/token-subscribe.go.html",
		"html/template/token-transfer.go.html",
		"html/template/yield.go.html")
	if err != nil {
		return err
	}

	keystore, err := bcgo.GetKeyDirectory(s.Root)
	if err != nil {
		return err
	}

	datastore := &conveygo.BCStore{
		Node:     node,
		Listener: s.Listener,
		KeyStore: keystore,
	}

	sessionstore := NewMemorySessionStore()

	var emailverifier EmailVerifier
	var emailwelcomer EmailWelcomer

	address, ok := os.LookupEnv("SMTP_ADDRESS")
	if !ok {
		log.Println("Missing SMTP_ADDRESS")
	} else {
		sender, ok := os.LookupEnv("SMTP_SENDER")
		if !ok {
			log.Println("Missing SMTP_SENDER")
		} else {
			emailverifier = NewSmtpEmailVerifier(address, sender, templates.Lookup("email-verification.go.html"))
			emailwelcomer = NewSmtpEmailWelcomer(address, sender, templates.Lookup("email-welcome.go.html"))
		}
	}

	var paymentprocessor PaymentProcessor

	secretKey, ok := os.LookupEnv("STRIPE_SECRET_KEY")
	if !ok {
		log.Println("Missing STRIPE_SECRET_KEY")
	} else {
		publishableKey, ok := os.LookupEnv("STRIPE_PUBLISHABLE_KEY")
		if !ok {
			log.Println("Missing STRIPE_PUBLISHABLE_KEY")
		} else {
			paymentprocessor = NewStripePaymentProcessor(secretKey, publishableKey, node, s.Listener)
		}
	}

	// Serve Web Requests
	mux := http.NewServeMux()
	mux.HandleFunc("/", netgo.StaticHandler("html/static"))
	mux.HandleFunc("/alias", aliasservergo.AliasHandler(aliases, s.Cache, s.Network, templates.Lookup("alias.go.html")))
	mux.HandleFunc("/block", bcnetgo.BlockHandler(s.Cache, s.Network, templates.Lookup("block.go.html")))
	mux.HandleFunc("/channel", bcnetgo.ChannelHandler(s.Cache, s.Network, templates.Lookup("channel.go.html")))
	mux.HandleFunc("/channels", bcnetgo.ChannelListHandler(s.Cache, s.Network, templates.Lookup("channel-list.go.html"), node.GetChannels))
	mux.HandleFunc("/keys", cryptogo.KeyShareHandler(make(cryptogo.KeyShareStore), 2*time.Minute))
	mux.HandleFunc("/account", AccountHandler(sessionstore, ledger, templates.Lookup("account.go.html")))
	// TODO(v2) mux.HandleFunc("/account-export", AccountExportHandler(sessionstore, templates.Lookup("account-export.go.html")))
	// TODO(v2) mux.HandleFunc("/account-import", AccountImportHandler(sessionstore, templates.Lookup("account-import.go.html")))
	mux.HandleFunc("/add-payment-method", AddPaymentMethodHandler(sessionstore, datastore, paymentprocessor, templates.Lookup("add-payment-method.go.html")))
	mux.HandleFunc("/best", BestHandler(sessionstore, datastore, templates.Lookup("best.go.html")))
	mux.HandleFunc("/compose", ComposeHandler(sessionstore, datastore, templates.Lookup("compose.go.html")))
	mux.HandleFunc("/conversation", ConversationHandler(sessionstore, datastore, templates.Lookup("conversation.go.html")))
	// TODO(v3) mux.HandleFunc("/digest", )
	mux.HandleFunc("/ledger", LedgerHandler(ledger, templates.Lookup("ledger.go.html")))
	mux.HandleFunc("/preview", PreviewHandler(sessionstore, datastore, ledger, templates.Lookup("preview.go.html")))
	mux.HandleFunc("/publish", PublishHandler(sessionstore, datastore, ledger, templates.Lookup("publish.go.html")))
	mux.HandleFunc("/recent", RecentHandler(sessionstore, datastore, templates.Lookup("recent.go.html")))
	mux.HandleFunc("/sign-in", SignInHandler(sessionstore, datastore, templates.Lookup("sign-in.go.html")))
	mux.HandleFunc("/sign-out", SignOutHandler(sessionstore, templates.Lookup("sign-out.go.html")))
	mux.HandleFunc("/sign-up", SignUpHandler(sessionstore, datastore, emailverifier, templates.Lookup("sign-up.go.html")))
	mux.HandleFunc("/sign-up-verification", SignUpVerificationHandler(sessionstore, datastore, paymentprocessor, emailwelcomer, templates.Lookup("sign-up-verification.go.html")))

	productId, ok := os.LookupEnv("PRODUCT_ID")
	if !ok {
		return errors.New("Missing PRODUCT_ID")
	}

	productIds := strings.Split(productId, ",")
	mux.HandleFunc("/token-purchase", TokenPurchaseHandler(sessionstore, datastore, paymentprocessor, ledger, node, templates.Lookup("token-purchase.go.html"), productIds))
	/* TODO(v3)
	planId := os.Getenv("PLAN_ID")
	if planId != "" {
		mux.HandleFunc("/token-subscribe", TokenSubscriptionHandler(sessionstore, datastore, paymentprocessor, node, templates.Lookup("token-subscribe.go.html"), productId, planId))
	}
	*/
	mux.HandleFunc("/token-transfer", TokenTransferHandler(sessionstore, datastore, ledger, aliases, transactions, node, s.Listener, templates.Lookup("token-transfer.go.html")))
	mux.HandleFunc("/stripe-webhook", bcnetgo.StripeWebhookHandler(NewStripeEventHandler(aliases, charges, transactions, node, s.Listener)))

	if bcgo.GetBooleanFlag("HTTPS") {
		// Redirect HTTP Requests to HTTPS
		go func() {
			if err := http.ListenAndServe(":80", http.HandlerFunc(netgo.HTTPSRedirect(node.Alias, map[string]bool{
				"/":                   true,
				"/account":            true,
				"/account-export":     true,
				"/account-import":     true,
				"/add-payment-method": true,
				"/alias":              true,
				"/best":               true,
				"/block":              true,
				"/channel":            true,
				"/channels":           true,
				"/compose":            true,
				"/conversation":       true,
				"/digest":             true,
				"/keys":               true,
				"/ledger":             true,
				"/preview":            true,
				"/recent":             true,
				"/sign-in":            true,
				"/sign-out":           true,
				"/sign-up":            true,
				"/token-purchase":     true,
				"/token-subscribe":    true,
				"/token-transfer":     true,
			}))); err != nil {
				log.Fatal(err)
			}
		}()
		// Serve HTTPS Requests
		config := &tls.Config{MinVersion: tls.VersionTLS10}
		server := &http.Server{Addr: ":443", Handler: mux, TLSConfig: config}
		log.Println("HTTPS Server listening on :443")
		return server.ListenAndServeTLS(path.Join(s.Cert, "fullchain.pem"), path.Join(s.Cert, "privkey.pem"))
	} else {
		log.Println("HTTP Server Listening on :80")
		return http.ListenAndServe(":80", mux)
	}
}

func (s *Server) Handle(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "init":
			PrintLegalese(os.Stdout)
			node, err := s.Init()
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Initialized")
			log.Println(node.Alias)
			publicKeyBytes, err := cryptogo.RSAPublicKeyToPKIXBytes(&node.Key.PublicKey)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(base64.RawURLEncoding.EncodeToString(publicKeyBytes))
		case "start":
			node, err := bcgo.GetNode(s.Root, s.Cache, s.Network)
			if err != nil {
				log.Println(err)
				return
			}
			if err := s.Start(node); err != nil {
				log.Println(err)
				return
			}
		default:
			log.Println("Cannot handle", args[0])
		}
	} else {
		PrintUsage(os.Stdout)
	}
}

func PrintUsage(output io.Writer) {
	fmt.Fprintln(output, "Convey Server Usage:")
	fmt.Fprintln(output, "\tconveyserver - display usage")
	fmt.Fprintln(output, "\tconveyserver init - initializes environment, generates key pair, and registers alias")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "\tconveyserver start - starts the server")
}

func PrintLegalese(output io.Writer) {
	fmt.Fprintln(output, "Convey Legalese:")
	fmt.Fprintln(output, "Convey is made available by Aletheia Ware LLC [https://aletheiaware.com] under the Terms of Service [https://aletheiaware.com/terms-of-service.html] and Privacy Policy [https://aletheiaware.com/privacy-policy.html].")
	fmt.Fprintln(output, "By continuing to use this software you agree to the Terms of Service, and Privacy Policy.")
}

func main() {
	rootDir, err := bcgo.GetRootDirectory()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Root Directory:", rootDir)

	logFile, err := bcgo.SetupLogging(rootDir)
	if err != nil {
		log.Println(err)
		return
	}
	defer logFile.Close()
	log.Println("Log File:", logFile.Name())

	certDir, err := bcgo.GetCertificateDirectory(rootDir)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Certificate Directory:", certDir)

	cacheDir, err := bcgo.GetCacheDirectory(rootDir)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Cache Directory:", cacheDir)

	cache, err := bcgo.NewFileCache(cacheDir)
	if err != nil {
		log.Println(err)
		return
	}

	peers, err := bcgo.GetPeers(rootDir)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Peers:", peers)

	network := bcgo.NewTCPNetwork(peers...)

	server := &Server{
		Root:     rootDir,
		Cert:     certDir,
		Cache:    cache,
		Network:  network,
		Listener: &bcgo.PrintingMiningListener{Output: os.Stdout},
	}

	server.Handle(os.Args[1:])
}
