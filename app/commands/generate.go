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
