// Copyright 2024 OpenPubkey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/awnumar/memguard"

	"github.com/openpubkey/openpubkey/client"
	"github.com/openpubkey/openpubkey/client/providers"
	"github.com/openpubkey/openpubkey/pktoken"
	"github.com/openpubkey/openpubkey/util"
	"golang.org/x/crypto/sha3"
)

// Variables for building our google provider
var (
	clientID = "992028499768-ce9juclb3vvckh23r83fjkmvf1lvjq18.apps.googleusercontent.com"
	// The clientSecret was intentionally checked in. It holds no power and is used for development. Do not report as a security issue
	clientSecret = "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" // Google requires a ClientSecret even if this a public OIDC App
	scopes       = []string{"openid profile email"}
	redirURIPort = "3000"
	callbackPath = "/login-callback"
	redirectURI  = fmt.Sprintf("http://localhost:%v%v", redirURIPort, callbackPath)

	// File names for when we save or load our pktoken and the corresponding signing key
	skFileName  = "key.pem"
	pktFileName = "pktoken.json"
)

func main() {
	// Safely terminate in case of an interrupt signal
	memguard.CatchInterrupt()

	// Purge the session when we return
	defer memguard.Purge()

	if len(os.Args) < 2 {
		fmt.Printf("OpenPubkey: command choices are login, sign, and cert")
		return
	}

	signGQ := true

	// Directory for saving data
	outputDir := "output/google"

	command := os.Args[1]
	switch command {
	case "login":
		if err := login(outputDir, signGQ); err != nil {
			fmt.Println("Error logging in:", err)
		} else {
			fmt.Println("Login successful!")
		}
	case "sign":
		message := "sign me!!"
		if err := sign(message, outputDir, signGQ); err != nil {
			fmt.Println("Failed to sign test message:", err)
		}
	default:
		fmt.Println("Unrecognized command:", command)
	}
}

func login(outputDir string, signGQ bool) error {
	opkClient, err := client.New(
		&providers.GoogleOp{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			RedirURIPort: redirURIPort,
			CallbackPath: callbackPath,
			RedirectURI:  redirectURI,
		},
		client.WithSignGQ(signGQ),
	)
	if err != nil {
		return err
	}

	pkt, err := opkClient.Auth(context.Background(),
		client.WithExtraClaim("extra", "yes"))
	if err != nil {
		return err
	}

	// Pretty print our json token
	pktJson, err := json.MarshalIndent(pkt, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(pktJson))

	// Save our signer and pktoken by writing them to a file
	return saveLogin(outputDir, opkClient.GetSigner().(*ecdsa.PrivateKey), pkt)
}

func sign(message string, outputDir string, signGq bool) error {
	signer, pkt, err := loadLogin(outputDir)
	if err != nil {
		return fmt.Errorf("failed to load client state: %w", err)
	}

	msgHashSum := sha3.Sum256([]byte(message))
	sig, err := signer.Sign(rand.Reader, msgHashSum[:], crypto.SHA256)
	if err != nil {
		return err
	}

	fmt.Println("Signed Message:", message)
	fmt.Println("Praise Sigma:", base64.StdEncoding.EncodeToString(sig))
	fmt.Println("Hash:", hex.EncodeToString(msgHashSum[:]))
	fmt.Println("Cert:")

	pktJson, err := json.Marshal(pkt)
	if err != nil {
		return err
	}

	// Pretty print our json token
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, pktJson, "", "  "); err != nil {
		return err
	}
	fmt.Println(prettyJSON.String())

	return nil
}

func saveLogin(outputDir string, sk *ecdsa.PrivateKey, pkt *pktoken.PKToken) error {
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return err
	}

	skFilePath := path.Join(outputDir, skFileName)
	if err := util.WriteSKFile(skFilePath, sk); err != nil {
		return err
	}

	pktFilePath := path.Join(outputDir, pktFileName)
	pktJson, err := json.Marshal(pkt)
	if err != nil {
		return err
	}
	return os.WriteFile(pktFilePath, pktJson, 0600)
}

func loadLogin(outputDir string) (crypto.Signer, *pktoken.PKToken, error) {
	skFilePath := path.Join(outputDir, skFileName)
	key, err := util.ReadSKFile(skFilePath)
	if err != nil {
		return nil, nil, err
	}

	pktFilePath := path.Join(outputDir, pktFileName)
	pktJson, err := os.ReadFile(pktFilePath)
	if err != nil {
		return nil, nil, err
	}

	var pkt *pktoken.PKToken
	if err := json.Unmarshal(pktJson, &pkt); err != nil {
		return nil, nil, err
	}

	return key, pkt, nil
}
