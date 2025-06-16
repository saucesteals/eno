package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/saucesteals/eno/extension"
	"github.com/saucesteals/eno/web"
)

func delete(ctx context.Context, capWeb *web.Web, card extension.PaymentCard) error {
	nameFilter := ask("Enter name filter (optional)")

	minDayAgeString := ask("Enter minimum age in days (optional)")
	if minDayAgeString == "" {
		minDayAgeString = "0"
	}

	minDayAge, err := strconv.Atoi(minDayAgeString)
	if err != nil {
		return fmt.Errorf("invalid minimum day age: %w", err)
	}

	cards := []web.ListedToken{}
	limit := 50
	for offset := 0; ; offset += limit {
		page, err := capWeb.ListTokens(ctx, card, nameFilter, offset, limit)
		if err != nil {
			return fmt.Errorf("list tokens: %w", err)
		}

		cards = append(cards, page.Entries...)
		if page.Count <= len(cards) {
			break
		}

	}

	fmt.Printf("Found %d cards\n", len(cards))
	for i, card := range cards {
		fmt.Printf("%d. %s (%s)\n", i+1, card.TokenName, card.TokenLastFour)
	}

	confirm := ask("Are you sure you want to delete these cards? (y/n)")
	if confirm != "y" {
		return nil
	}

	cardLastFour := card.CardNumber[len(card.CardNumber)-4:]
	for _, token := range cards {
		tokenCreatedAt, err := time.ParseInLocation("2006-01-02T15:04:05", token.TokenCreatedTimestamp, time.UTC)
		if err != nil {
			return fmt.Errorf("invalid token created at: %w", err)
		}

		dayAge := time.Since(tokenCreatedAt).Hours() / 24
		if dayAge < float64(minDayAge) {
			fmt.Printf("Skipping %s (%.2f days old)\n", token.TokenName, dayAge)
			continue
		}

		fmt.Printf("Deleting %s... ", token.TokenName)

		err = capWeb.UpdateToken(ctx, web.UpdateTokenPayload{
			AllowAuthorizations: true,
			CardLastFour:        cardLastFour,
			CardName:            card.ProductDescription,
			CardReferenceID:     card.CardReferenceID,
			IsDeleted:           true,
			MdxID:               token.MdxInfo.MdxID,
			MdxURLID:            token.MdxInfo.MdxURLID,
			TokenDuration:       nil,
			TokenLastFour:       token.TokenLastFour,
			TokenName:           token.TokenName,
			TokenReferenceID:    token.TokenReferenceID,
		})
		if err != nil {
			log.Error("Failed to delete token", "error", err)
			continue
		}

		fmt.Println("Done")
	}

	return nil
}
