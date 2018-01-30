package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/smartcontractkit/chainlink/logger"
	"github.com/smartcontractkit/chainlink/store"
	"golang.org/x/crypto/ssh/terminal"
)

type Authenticator interface {
	Authenticate(*store.Store, string)
}

type TerminalAuthenticator struct {
	Prompter Prompter
	Exiter   func(int)
}

// Authenticate checks to see if there are accounts present in
// the KeyStore, and if there are none, a new account will be created
// by prompting for a password. If there are accounts present, the
// account which is unlocked by the given password will be used.
func (auth TerminalAuthenticator) Authenticate(store *store.Store, pwd string) {
	if len(pwd) != 0 {
		auth.authenticateWithPwd(store, pwd)
	} else {
		auth.authenticationPrompt(store)
	}
}

func (auth TerminalAuthenticator) authenticationPrompt(store *store.Store) {
	if store.KeyStore.HasAccounts() {
		auth.promptAndCheckPassword(store)
	} else {
		auth.createAccount(store)
	}
}

func (auth TerminalAuthenticator) authenticateWithPwd(store *store.Store, pwd string) {
	if !store.KeyStore.HasAccounts() {
		fmt.Println("Cannot authenticate with password because there are no accounts")
		auth.Exiter(1)
	} else if err := checkPassword(store, pwd); err != nil {
		auth.Exiter(1)
	}
}

func checkPassword(store *store.Store, phrase string) error {
	if err := store.KeyStore.Unlock(phrase); err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		printGreeting()
		return nil
	}
}

func (auth TerminalAuthenticator) promptAndCheckPassword(store *store.Store) {
	for {
		phrase := auth.Prompter.Prompt("Enter Password:")
		if checkPassword(store, phrase) == nil {
			break
		}
	}
}

func (auth TerminalAuthenticator) createAccount(store *store.Store) {
	for {
		phrase := auth.Prompter.Prompt("New Password: ")
		phraseConfirmation := auth.Prompter.Prompt("Confirm Password: ")
		if phrase == phraseConfirmation {
			_, err := store.KeyStore.NewAccount(phrase)
			if err != nil {
				logger.Fatal(err)
			}
			printGreeting()
			break
		} else {
			fmt.Println("Passwords don't match. Please try again.")
		}
	}
}

type Prompter interface {
	Prompt(string) string
}

type PasswordPrompter struct{}

func (pp PasswordPrompter) Prompt(prompt string) string {
	var rval string
	withTerminalResetter(func() {
		fmt.Print(prompt)
		bytePwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Println()
		rval = string(bytePwd)
	})
	return rval
}

// Explicitly reset terminal state in the event of a signal (CTRL+C)
// to ensure typed characters are echoed in terminal:
// https://groups.google.com/forum/#!topic/Golang-nuts/kTVAbtee9UA
func withTerminalResetter(f func()) {
	initialTermState, err := terminal.GetState(syscall.Stdin)
	if err != nil {
		logger.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		terminal.Restore(syscall.Stdin, initialTermState)
		os.Exit(1)
	}()

	f()
	signal.Stop(c)
}

func printGreeting() {
	fmt.Println(`
     _/_/_/  _/                  _/            _/        _/            _/
  _/        _/_/_/      _/_/_/      _/_/_/    _/            _/_/_/    _/  _/
 _/        _/    _/  _/    _/  _/  _/    _/  _/        _/  _/    _/  _/_/
_/        _/    _/  _/    _/  _/  _/    _/  _/        _/  _/    _/  _/  _/
 _/_/_/  _/    _/    _/_/_/  _/  _/    _/  _/_/_/_/  _/  _/    _/  _/    _/
`)
}