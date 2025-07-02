package main

import (
	"context"
	"fmt"

	"github.com/saucesteals/eno/extension"
	"github.com/saucesteals/eno/web"
)

func list(ctx context.Context, capWeb *web.Web, card extension.PaymentCard) error {
	nameFilter := ask("Enter name filter (optional)")

	limit := 50
	var total int
	for offset := 0; ; offset += 1 {
		page, err := capWeb.ListTokens(ctx, card, nameFilter, offset, limit)
		if err != nil {
			return fmt.Errorf("list tokens: %w", err)
		}

		for _, token := range page.Entries {
			fmt.Printf("- %q on %q\n", token.TokenName, token.MdxInfo.MerchantURL)
		}

		total += page.Count
		if page.Count <= total {
			break
		}
	}

	log.Info("Found cards", "count", total)

	return nil
}
