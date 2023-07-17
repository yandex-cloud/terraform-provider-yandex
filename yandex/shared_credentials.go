package yandex

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Profile = string
type RawCredentials = map[string]string

// A SharedCredentialsProvider retrieves credentials for the selected profile from the given file.
type SharedCredentialsProvider struct {
	// Path to the shared credentials file.
	Filename string

	// YC Profile to extract credentials from the shared credentials file.
	Profile string
}

type SharedCredentials struct {
	StorageAccessKey string
	StorageSecretKey string
}

// Retrieve reads and extracts the credentials for the selected profile from the given file.
func (p *SharedCredentialsProvider) Retrieve() (*SharedCredentials, error) {
	rawCredentialsByProfiles, err := parse(p.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shared credentials file, error: \"%w\"", err)
	}

	rawCredentials, ok := rawCredentialsByProfiles[p.Profile]
	if !ok {
		return nil, fmt.Errorf("not found shared credentials for `%v` profile", p.Profile)
	}

	credentials := SharedCredentials{}
	if val, ok := rawCredentials["storage_access_key"]; ok {
		credentials.StorageAccessKey = val
	}

	if val, ok := rawCredentials["storage_secret_key"]; ok {
		credentials.StorageSecretKey = val
	}

	return &credentials, nil
}

func (p *SharedCredentials) HasStorageAccessKeys() bool {
	return p.StorageAccessKey != "" && p.StorageSecretKey != ""
}

func parse(filename string) (map[Profile]RawCredentials, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	credentialsByProfiles, err := parseLines(scanner)
	if err != nil {
		return nil, err
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return credentialsByProfiles, nil
}

func parseLines(scanner *bufio.Scanner) (map[Profile]RawCredentials, error) {
	credentialsByProfiles := make(map[Profile]RawCredentials)

	profileRegex := regexp.MustCompile(`^\[(.+)\]$`)
	keyValuePairRegex := regexp.MustCompile(`^([ \w_]+)=(.+)$`)

	var currentProfile string
	var currentProfileCredentials RawCredentials
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if profileRegex.MatchString(line) {
			if currentProfile != "" {
				credentialsByProfiles[currentProfile] = currentProfileCredentials
			}

			currentProfile = profileRegex.FindStringSubmatch(line)[1]
			currentProfileCredentials = make(map[string]string)
		} else if keyValuePairRegex.MatchString(line) {
			keyValue := keyValuePairRegex.FindStringSubmatch(line)
			key := strings.TrimSpace(keyValue[1])
			value := strings.TrimSpace(keyValue[2])
			if currentProfile == "" {
				return nil, fmt.Errorf("key `%v` does not have a profile", key)
			}
			if _, exists := currentProfileCredentials[key]; exists {
				return nil, fmt.Errorf("key `%v` has multiple values", key)
			}
			currentProfileCredentials[key] = value
		} else {
			return nil, fmt.Errorf("unsupported shared credentials file format, line `%v` does not match "+
				"either `[{profile}]` or `{key}={value}` patterns", line)
		}
	}

	if currentProfile != "" {
		credentialsByProfiles[currentProfile] = currentProfileCredentials
	}

	return credentialsByProfiles, nil
}
