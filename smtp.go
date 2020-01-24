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
	"bytes"
	"github.com/AletheiaWareLLC/cryptogo"
	"html/template"
	"log"
	"net/smtp"
)

const (
	ERROR_INCORRECT_EMAIL_VERIFICATION = "Incorrect Email Verification Code"
	VERIFICATION_CODE_LENGTH           = 6
)

/*
data := struct {
	From string
	To string
	Subject string
	Content string
}{
	From: from,
	To: to,
	Subject: subject,
	Content: content,
}
if err := SetEmail(address, from, to, template, data); err != nil {
	return "", err
}
*/

func SetEmail(address, from, to string, template *template.Template, data interface{}) error {
	var buffer bytes.Buffer
	if err := template.Execute(&buffer, data); err != nil {
		log.Println(err)
		return err
	}
	return smtp.SendMail(address, nil, from, []string{to}, buffer.Bytes())
}

type SmtpEmailVerifier struct {
	Address  string
	Sender   string
	Template *template.Template
}

func NewSmtpEmailVerifier(address, sender string, template *template.Template) *SmtpEmailVerifier {
	return &SmtpEmailVerifier{
		Address:  address,
		Sender:   sender,
		Template: template,
	}
}

func (v SmtpEmailVerifier) VerifyEmail(email string) (string, error) {
	log.Println("Verifying Email", email)
	code, err := cryptogo.RandomString(VERIFICATION_CODE_LENGTH)
	if err != nil {
		return "", err
	}
	log.Println("Verification Code", code)
	data := struct {
		From      string
		To        string
		Challenge string
	}{
		From:      v.Sender,
		To:        email,
		Challenge: code,
	}
	if err := SetEmail(v.Address, v.Sender, email, v.Template, data); err != nil {
		return "", err
	}
	return code, nil
}

type SmtpEmailWelcomer struct {
	Address  string
	Sender   string
	Template *template.Template
}

func NewSmtpEmailWelcomer(address, sender string, template *template.Template) *SmtpEmailWelcomer {
	return &SmtpEmailWelcomer{
		Address:  address,
		Sender:   sender,
		Template: template,
	}
}

func (v SmtpEmailWelcomer) WelcomeEmail(alias, email string) error {
	log.Println("Welcoming Email", email)
	data := struct {
		From  string
		To    string
		Alias string
	}{
		From:  v.Sender,
		To:    email,
		Alias: alias,
	}
	if err := SetEmail(v.Address, v.Sender, email, v.Template, data); err != nil {
		return err
	}
	return nil
}
