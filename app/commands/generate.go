// Copyright Jeevanandam M. (https://github.com/jeevatkm, jeeva@myjeeva.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"aahframe.work/console"
	"aahframe.work/essentials"
)

// Generate command provide an easy option for THUMBAI users to create random secure key using `crypto/rand`.
// Mainly it provided for Security Config (Session and Anti-CRSF sign and encrypt key).
var Generate = console.Command{
	Name:    "generate",
	Aliases: []string{"g"},
	Usage:   "Generates random secure keys using `crypto/rand` for security config.",
	Description: `Generate command provide an easy option for THUMBAI users to create random secure key using 'crypto/rand'.
	Mainly it provided for Security Configs (Session & Anti-CRSF) encryption key for AES-256 & signature key for HMAC.

	To know more about available 'generate' sub commands:
		thumbai help generate

	To know more about individual sub-commands details:
		thumbai generate help <sub-command-name>`,
	Subcommands: []console.Command{
		{
			Name:    "securekeys",
			Aliases: []string{"sk"},
			Usage:   "Generates secure config keys for session and anti-csrf.",
			Description: `Generates secure config keys for session and anti-csrf (AES-256 & HMAC).

	Examples:
		thumbai generate securekeys`,
			Action: generateSecureKeysAction,
		},
	},
}

const secureKeysCfgTmplStr = `
{{ .SectionName }} {
  sign_key = "{{ .SignKey }}"
  enc_key = "{{ .EncKey }}"
}
`

func generateSecureKeysAction(c *console.Context) error {
	cl := log.New(os.Stderr, "", 0)
	fmt.Println("Generating secure keys for Session & Anti-CSRF using `crypto/rand`.")
	fmt.Println("Add below config into section 'security { ... }' on file 'thumbai.conf'.")
	tmpl, err := template.New("secure_keys").Parse(secureKeysCfgTmplStr)
	if err != nil {
		cl.Fatalln(err)
	}
	_ = tmpl.Execute(os.Stdout, map[string]interface{}{
		"SectionName": "session",
		"SignKey":     ess.SecureRandomString(64),
		"EncKey":      ess.SecureRandomString(32),
	})
	_ = tmpl.Execute(os.Stdout, map[string]interface{}{
		"SectionName": "anti_csrf",
		"SignKey":     ess.SecureRandomString(64),
		"EncKey":      ess.SecureRandomString(32),
	})
	return nil
}
