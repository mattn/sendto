package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gophergala2016/sendto/client"
)

const (
	v = "0.1"
)

func main() {
	command := ""
	args := os.Args[1:] // remove app path from args

	// We expect either a username or a subcommand and then a set of files in args
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	// Load our configuration
	err := client.LoadConfig()
	if err != nil {
		log.Fatalf("Sorry, an error occurred:\n\t%s", err)
	}

	switch command {
	case "encrypt", "e":
		err = Encrypt(args)
	case "decrypt", "d":
		err = Decrypt(args)
	case "identity", "i":
		err = Identity(args)
	case "version", "v":
		Version()
	case "help", "h":
		Help()
	default:
		// Default action is to send to (if we have a username and files)
		if len(args) > 0 {
			err = SendTo(command, args)
		} else {
			Help()
		}
	}

	if err != nil {
		log.Fatalf("Sorry, an error occurred:\n\t%s", err)
	}
}

// Version prints the version of this app
func Version() {
	fmt.Printf("\n\t-----\n\tSend to client - version:%s\n\t-----\n", v)
}

// Usage returns standard usage as a string
func Usage() string {
	return fmt.Sprintf("\tUsage: sendto kennygrant [files] - send files to the username kennygrant\n")
}

// Help prints the usage and commands
func Help() {
	Version()
	fmt.Printf(Usage())
	fmt.Printf("\t-----\n")
	fmt.Printf("\tCommands:\n")
	fmt.Printf("\tsendto version - display version\n")
	fmt.Printf("\tsendto [username] [files] - encrypt files for a given user\n")
	fmt.Printf("\tsendto encrypt [file] - encrypt a file\n")
	//	fmt.Printf("\tsendto decrypt [file] - decrypt a file\n")
	fmt.Printf("\tsendto identity [name] - sets default sender identity\n\n")
}

// Decrypt files specified, using the user's private key
// TODO: to support decryption we'd need access to private keys, perhaps leave this for hackathon
func Decrypt(args []string) error {
	log.Printf("Sorry, this client does not yet support decrypt")

	return nil
}

// Encrypt the files specified
func Encrypt(args []string) error {

	log.Printf("Sorry, this client does not yet support encryption")
	return nil
}

// SendTo sends files held in args to recipient
func SendTo(recipient string, args []string) error {

	// We expect at least 1 file to send
	if len(args) < 1 {
		return fmt.Errorf("Not enough arguments - %s", Usage())
	}

	// Notify the user that we're starting to send
	fmt.Printf("Sending %d %s to %s as %s...\n", len(args), filesString(len(args)), recipient, client.Config["sender"])

	// Fetch the recipient's key (from disk or server)

	// For the moment as a test, use keybase.io, should be using our server
	keyURL := fmt.Sprintf(client.Config["keyserver"], recipient)
	keyPath, err := client.LoadKey(recipient, keyURL)
	if err != nil {
		// Warn user in a nicer way here that key could not be found
		return fmt.Errorf("Failed to find key:%s", err)
	}
	fmt.Printf("Loaded key for %s:\n%s\n", recipient, keyPath)

	// Zip and Encrypt our arguments (files or folders) using key
	dataPath, err := client.EncryptFiles(args, recipient, keyPath)
	if err != nil {
		return err
	}

	// Send the file to the recipient on the server
	postURL := fmt.Sprintf("%s/files/create", client.Config["server"])

	fmt.Printf("Sending files for %s to %s\n", recipient, postURL)

	err = client.PostData(client.Config["sender"], recipient, dataPath, postURL)
	if err != nil {
		return err
	}

	return nil
}

// Identity sets the default sender identity (as opposed to username)
func Identity(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Identity command requires a sender name")
	}

	identity := args[0]
	client.Config["sender"] = identity

	fmt.Printf("Setting sender identity to:%s\n", identity)

	return client.SaveConfig()
}

// Perhaps also allow setting default server?

// Return a nicely formatted string for the word files
func filesString(i int) string {
	if i > 1 {
		return "files"
	}
	return "file"
}
