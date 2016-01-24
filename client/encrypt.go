package client

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// LoadKey loads the key associated with this username,
// first by loooking at ~/.sendto/users/recipient/key.pub
// or if that fails by fetching it from the internet and saving at that location
// it returns the path of the downloaded key file
func LoadKey(recipient string, url string) (string, error) {
	fmt.Printf("Loading key for %s...\n", recipient)

	// Define a local keypath of users/name/key.pub
	keyPath := filepath.Join(configPath(), "users", recipient, "key.pub")

	// Check if the key file exists at ~/.sendto/users/recipient/key.pub
	if !fileExists(keyPath) {
		// Make the user directory
		createFolder(filepath.Join("users", recipient))

		// Fetch the key from our server
		err := DownloadData(url, keyPath)
		if err != nil {
			return "", err
		}

		// Tell user we fetched key
		fmt.Printf("Fetched key for user:%s from:%s\n", recipient, url)

		// Print the key for the user as we have fetched it for the first time?
		/*
		   key, err := ioutil.ReadFile(keyPath)
		   if err != nil {
		     return "", err
		   }
		*/
	}

	return keyPath, nil
}

// EncryptFiles zips then encrypts our arguments (files or folders) using a public key
func EncryptFiles(args []string, recipient string, keyPath string) (string, error) {

	// First open and parse recipient key
	publicKey, err := ParsePublicKey(keyPath)
	if err != nil {
		return "", err
	}

	fmt.Printf("Using key: %x\n", publicKey.PrimaryKey.Fingerprint)

	// Make the user files directory
	createFolder(filepath.Join("files", recipient))

	// Should caller set the filename required for the zip?
	// hash of username + time or something? Doesn't really matter but should be unique I guess
	// it doesn't have to be unique though... just use caller name for now
	name := path.Clean(path.Base(args[0]))
	outPath := filepath.Join(configPath(), "files", recipient, fmt.Sprintf("%s.zip.gpg", name))
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Create encryption writer
	hints := &openpgp.FileHints{IsBinary: true, FileName: fmt.Sprintf("%s.zip", name), ModTime: time.Now()}
	pgpWriter, err := openpgp.Encrypt(out, []*openpgp.Entity{publicKey}, nil, hints, nil)
	if err != nil {
		return "", err
	}

	// Now create a zipwriter, which writes to this pgpWriter
	zipWriter := zip.NewWriter(pgpWriter)

	// Add the files/folders from our args
	for _, arg := range args {

		// For each argument, walk the file path adding files to our zip
		err := filepath.Walk(arg, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()

			// Support unicode filenames by default
			h := &zip.FileHeader{Name: p, Method: zip.Deflate, Flags: 0x800}
			z, err := zipWriter.CreateHeader(h)
			// Doesn't support unicode file names?
			// z, err := zipWriter.Create(p)
			if err != nil {
				return err
			}
			io.Copy(z, f)
			zipWriter.Flush()
			return nil
		})
		if err != nil {
			return "", err
		}

	}
	err = zipWriter.Flush()
	if err != nil {
		return "", err
	}
	err = zipWriter.Close()
	if err != nil {
		return "", err
	}

	// close the encPipe to finish the process
	err = pgpWriter.Close()

	return outPath, err
}

// ParsePublicKey parses the given public key file
func ParsePublicKey(keyPath string) (*openpgp.Entity, error) {
	f, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse our key
	key, err := armor.Decode(f)
	if err != nil {
		return nil, err
	}
	if key.Type != openpgp.PublicKeyType {
		return nil, fmt.Errorf("Key of wrong type:%s", key.Type)
	}
	r := packet.NewReader(key.Body)
	to, err := openpgp.ReadEntity(r)
	if err != nil {
		return nil, err
	}

	return to, nil
}

// DecryptFiles decrypts then unzips a file using a private key
// and returns the path of the resulting file/folder on success
// zip step should be optional TODO
func DecryptFiles(p string, key string) (string, error) {

	return "", nil
}
