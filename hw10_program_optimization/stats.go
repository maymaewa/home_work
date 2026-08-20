package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

var json = jsoniter.ConfigFastest

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	suffix := "." + strings.ToLower(domain)

	for scanner.Scan() {
		var user struct {
			Email string `json:"Email"` //nolint:tagliatelle
		}

		if err := json.Unmarshal(scanner.Bytes(), &user); err != nil {
			return nil, fmt.Errorf("unmarshal user: %w", err)
		}

		email := strings.ToLower(user.Email)

		if !strings.HasSuffix(email, suffix) {
			continue
		}

		_, domainName, ok := strings.Cut(email, "@")
		if !ok {
			continue
		}

		result[domainName]++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan users: %w", err)
	}

	return result, nil
}
