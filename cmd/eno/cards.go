package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/saucesteals/eno/api"
	"github.com/saucesteals/eno/extension"
)

type CardWriter struct {
	f    *os.File
	path string
}

func cleanName(name string) string {
	return strings.ToLower(
		strings.ReplaceAll(name, " ", "_"),
	)
}

func NewCardWriter(profile *Profile, card extension.PaymentCard, suffix string) (*CardWriter, error) {
	dir, err := profile.GetDirectory(
		"cards",
		cleanName(card.ProductDescription),
	)
	if err != nil {
		return nil, err
	}

	t := time.Now().Format("2006_01_02_15_04_05")
	fileName := path.Join(dir, fmt.Sprintf("%s_%s.csv", t, cleanName(suffix)))
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &CardWriter{f: f, path: fileName}, nil
}

func (w *CardWriter) GetPath() string {
	return w.path
}

func (w *CardWriter) Write(card api.Token) error {
	expirationParts := strings.Split(card.ExpirationDate, "/")
	if len(expirationParts) != 2 {
		return fmt.Errorf("invalid expiration date: %s", card.ExpirationDate)
	}

	expMonth, err := strconv.Atoi(strings.TrimPrefix(expirationParts[0], "0"))
	if err != nil {
		return err
	}

	expYear, err := strconv.Atoi(expirationParts[1])
	if err != nil {
		return err
	}

	if expYear < 100 {
		expYear += 2000
	}

	line := fmt.Sprintf("%s,%d,%d,%s", card.Token, expMonth, expYear, card.Cvv)
	_, err = w.f.Write([]byte(line + "\n"))
	return err
}

func (w *CardWriter) Close() error {
	return w.f.Close()
}
