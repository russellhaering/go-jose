/*-
 * Copyright 2014 Square Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/square/go-jose"
	"io/ioutil"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "jose-util"
	app.Usage = "command-line utility to deal with JOSE objects"
	app.Version = "0.0.1"
	app.Author = ""
	app.Email = ""

	app.Commands = []cli.Command{
		{
			Name:  "encrypt",
			Usage: "encrypt a plaintext",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key, k",
					Usage: "Path to key file (PEM/DER)",
				},
				cli.StringFlag{
					Name:  "input, in",
					Usage: "Path to input file (stdin if missing)",
				},
				cli.StringFlag{
					Name:  "output, out",
					Usage: "Path to output file (stdout if missing)",
				},
				cli.StringFlag{
					Name:  "algorithm, alg",
					Usage: "Key management algorithm (e.g. RSA-OAEP)",
				},
				cli.StringFlag{
					Name:  "encryption, enc",
					Usage: "Content encryption algorithm (e.g. A128GCM)",
				},
				cli.BoolFlag{
					Name:  "full, f",
					Usage: "Use full serialization format (instead of compact)",
				},
			},
			Action: func(c *cli.Context) {
				keyBytes, err := ioutil.ReadFile(requiredFlag(c, "key"))
				exitOnError(err, "unable to read key file")

				pub, err := jose.LoadPublicKey(keyBytes)
				exitOnError(err, "unable to read public key")

				alg := jose.KeyAlgorithm(requiredFlag(c, "alg"))
				enc := jose.ContentEncryption(requiredFlag(c, "enc"))

				crypter, err := jose.NewEncrypter(alg, enc, pub)
				exitOnError(err, "unable to instantiate encrypter")

				obj, err := crypter.Encrypt(readInput(c.String("input")))
				exitOnError(err, "unable to encrypt")

				var msg string
				if c.Bool("full") {
					msg = obj.FullSerialize()
				} else {
					msg, err = obj.CompactSerialize()
					exitOnError(err, "unable to serialize message")
				}

				writeOutput(c.String("output"), []byte(msg))
			},
		},
		{
			Name:  "decrypt",
			Usage: "decrypt a plaintext",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key, k",
					Usage: "Path to key file (PEM/DER)",
				},
				cli.StringFlag{
					Name:  "input, in",
					Usage: "Path to input file (stdin if missing)",
				},
				cli.StringFlag{
					Name:  "output, out",
					Usage: "Path to output file (stdout if missing)",
				},
			},
			Action: func(c *cli.Context) {
				keyBytes, err := ioutil.ReadFile(requiredFlag(c, "key"))
				exitOnError(err, "unable to read private key")

				priv, err := jose.LoadPrivateKey(keyBytes)
				exitOnError(err, "unable to read private key")

				obj, err := jose.ParseEncrypted(string(readInput(c.String("input"))))
				exitOnError(err, "unable to parse message")

				plaintext, err := obj.Decrypt(priv)
				exitOnError(err, "unable to decrypt message")

				writeOutput(c.String("output"), plaintext)
			},
		},
	}

	err := app.Run(os.Args)
	exitOnError(err, "unable to run application")
}

// Retrieve value of a required flag
func requiredFlag(c *cli.Context, flag string) string {
	value := c.String(flag)
	if value == "" {
		fmt.Fprintf(os.Stderr, "missing required flag --%s\n", flag)
		os.Exit(1)
	}
	return value
}

// Exit and print error message if we encountered a problem
func exitOnError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
		os.Exit(1)
	}
}

// Read input from file or stdin
func readInput(path string) []byte {
	var bytes []byte
	var err error

	if path != "" {
		bytes, err = ioutil.ReadFile(path)
	} else {
		bytes, err = ioutil.ReadAll(os.Stdin)
	}

	exitOnError(err, "unable to read input")
	return bytes
}

// Write output to file or stdin
func writeOutput(path string, data []byte) {
	var err error

	if path != "" {
		err = ioutil.WriteFile(path, data, 0644)
	} else {
		_, err = os.Stdout.Write(data)
	}

	exitOnError(err, "unable to write output")
}
