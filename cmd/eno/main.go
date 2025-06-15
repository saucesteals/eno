package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	http "github.com/saucesteals/fhttp"

	"github.com/saucesteals/eno/api"
	"github.com/saucesteals/eno/extension"
	"github.com/saucesteals/eno/web"
)

var (
	log = slog.New(tint.NewHandler(colorable.NewColorable(os.Stdout), &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.DateOnly,
	}))
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Unexpected error", "error", r)
		}

		log.Info("Press ENTER to exit...")
		fmt.Scanln()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	profile, err := loadProfile()
	if err != nil {
		log.Error("Load profile", "error", err)
		return
	}

	credentials, err := profile.Credentials.Get()
	if err != nil {
		log.Error("Get credentials", "error", err)
		return
	}

	browserBin := os.Getenv("ENO_BROWSER_BINARY")
	if browserBin == "" {
		browserBin = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		if runtime.GOOS == "windows" {
			browserBin = "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
		}
	}

	if _, err := os.Stat(browserBin); err != nil {
		if os.IsNotExist(err) {
			log.Error("Please install Google Chrome or set a custom binary with the ENO_BROWSER_BINARY environment variable")
			return
		}

		log.Error("Check browser binary", "error", err)
		return
	}

	userDataDir, err := profile.GetDirectory("user_data")
	if err != nil {
		log.Error("Get user data directory", "error", err)
		return
	}

	capApi, err := api.New(api.Options{
		Logger:              log,
		Credentials:         credentials,
		BrowserUserDataPath: userDataDir,
		BrowserBinary:       browserBin,
	})
	if err != nil {
		log.Error("New API", "error", err)
		return
	}

	device, err := profile.Device.Get()
	if err != nil {
		if !errors.Is(err, ErrResourceMissing) {
			log.Error("Get device", "error", err)
			return
		}

		device = extension.GenerateDevice(capApi.GetUserAgent())
		err = profile.Device.Set(device)
		if err != nil {
			log.Error("Set device", "error", err)
			return
		}
	}

	cookies, err := profile.Cookies.Get()
	if err != nil {
		if !errors.Is(err, ErrResourceMissing) {
			log.Error("Get cookies", "error", err)
			return
		}

		cookies = []*http.Cookie{}
		err = profile.Cookies.Set(cookies)
		if err != nil {
			log.Error("Set cookies", "error", err)
			return
		}
	}

	capApi.SetCookies(cookies)

	capWeb := web.New(capApi)

	capExt, err := extension.New(capApi, device)
	if err != nil {
		log.Error("New extension", "error", err)
		return
	}

	err = login(ctx, profile, capExt, capApi)
	if err != nil {
		log.Error("Login", "error", err)
		return
	}

	err = profile.Cookies.Set(capApi.GetCookies())
	if err != nil {
		log.Error("Save cookies after login", "error", err)
		return
	}

	cards, err := capExt.GetPaymentCards(ctx)
	if err != nil {
		log.Error("Get payment cards", "error", err)
		return
	}

	cardNumbers := []string{}
	for i, card := range cards {
		cardNumbers = append(cardNumbers, fmt.Sprintf("%d. %s", i+1, card.CardNumber))
	}

	fmt.Printf("Cards:\n%s\n", strings.Join(cardNumbers, "\n"))
	var card extension.PaymentCard
	for {
		cardIndex, err := strconv.Atoi(ask("Enter card number"))
		if err != nil {
			log.Error("Invalid card number", "error", err)
			continue
		}

		if cardIndex < 1 || cardIndex > len(cards) {
			log.Error("Invalid card number", "card", cardIndex)
			continue
		}

		card = cards[cardIndex-1]
		break
	}

	commands := []string{
		"create",
		"list",
		"delete",
		"exit",
	}

	for {
		var command string
		for {
			command = ask(fmt.Sprintf("Enter command (%s)", strings.Join(commands, ", ")))
			if slices.Contains(commands, command) {
				break
			}

			log.Error("Invalid command", "command", command)
		}

		switch command {
		case "exit":
			log.Info("Exiting...")
			os.Exit(0)
			return
		case "list":
			nameFilter := ask("Enter name filter (optional)")

			limit := 50
			var total int
			for offset := 0; ; offset += limit {
				page, err := capWeb.ListTokens(ctx, card, nameFilter, offset, limit)
				if err != nil {
					log.Error("List tokens", "error", err)
					return
				}

				for _, token := range page.Entries {
					fmt.Printf("- %q on %q\n", token.TokenName, token.MdxInfo.MerchantURL)
				}

				total += page.Count
				if page.Count <= total {
					break
				}
			}
		case "delete":
			nameFilter := ask("Enter name filter (optional)")

			minDayAgeString := ask("Enter minimum age in days (optional)")
			if minDayAgeString == "" {
				minDayAgeString = "0"
			}

			minDayAge, err := strconv.Atoi(minDayAgeString)
			if err != nil {
				log.Error("Invalid minimum day age", "error", err)
				return
			}

			var total int
			limit := 50
			for offset := 0; ; offset += limit {
				page, err := capWeb.ListTokens(ctx, card, nameFilter, offset, limit)
				if err != nil {
					log.Error("List tokens", "error", err)
					return
				}

				cardLastFour := card.CardNumber[len(card.CardNumber)-4:]

				for _, token := range page.Entries {
					tokenCreatedAt, err := time.ParseInLocation("2006-01-02T15:04:05", token.TokenCreatedTimestamp, time.UTC)
					if err != nil {
						log.Error("Invalid token created at", "error", err)
						return
					}

					dayAge := time.Since(tokenCreatedAt).Hours() / 24
					if dayAge < float64(minDayAge) {
						fmt.Printf("Skipping %s (%d days old)\n", token.TokenName, int(dayAge))
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

				total += page.Count
				if page.Count <= total {
					break
				}
			}
		case "create":
			count, err := strconv.Atoi(ask("Enter number of cards to create"))
			if err != nil {
				log.Error("Invalid number of cards", "error", err)
				return
			}

			merchantUrl := ask("Enter merchant URL (e.g. www.google.com)")
			merchant, err := capExt.DataSourceSearch(ctx, merchantUrl)
			if err != nil {
				log.Error("Failed to search for merchant", "error", err)
				return
			}

			w, err := NewCardWriter(profile, card, merchant)
			if err != nil {
				log.Error("New card writer", "error", err)
				return
			}
			defer w.Close()

			delay := time.Second * 5
			maxTries := 3
			for i := range count {
				var token extension.Token
				for j := range maxTries {
					token, err = capExt.CreateToken(ctx, card, merchant, fmt.Sprintf("%s Token %d", merchant.Name, i+1))
					if err != nil {
						log.Error("Failed to create token", "error", err)
						if j == maxTries-1 {
							log.Error("Failed to create token after max tries", "tries", maxTries)
							return
						}

						if strings.Contains(err.Error(), "status code: 429") {
							log.Info("Rate limited, sleeping for 2 minutes")
							time.Sleep(time.Minute * 2)
							continue
						}

						continue
					}

					break
				}

				log.Info("Created token", "token", token.Token)

				err = w.Write(&token)
				if err != nil {
					log.Error("Failed to write token", "error", err)
					return
				}

				if i < count-1 {
					time.Sleep(delay)
				}
			}

			log.Info("Created cards", "count", count, "path", w.GetPath())
		}
	}
}

func ask(prompt string) string {
	fmt.Printf("[?] %s: ", prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

func loadProfile() (*Profile, error) {
	username := ask("Enter username")

	profile, err := ImportProfile(username)
	if err != nil {
		return nil, fmt.Errorf("import profile: %w", err)
	}

	credentials, err := profile.Credentials.Get()
	if err != nil {
		if !errors.Is(err, ErrResourceMissing) {
			return nil, fmt.Errorf("get credentials: %w", err)
		}

		credentials = api.Credentials{
			Username: username,
			Password: ask("Enter password"),
		}

		err = profile.Credentials.Set(credentials)
		if err != nil {
			return nil, fmt.Errorf("set credentials: %w", err)
		}
	}

	return profile, nil
}

func login(ctx context.Context, profile *Profile, capExt *extension.Extension, capApi *api.API) error {
	express, err := profile.Express.Get()
	if err != nil {
		if !errors.Is(err, ErrResourceMissing) {
			return fmt.Errorf("get express: %w", err)
		}

		express = extension.ExpressEnrollment{}
		err = profile.Express.Set(express)
		if err != nil {
			return fmt.Errorf("set express: %w", err)
		}
	}

	for {
		session, err := capExt.GetSession(ctx)
		if err != nil {
			err = capApi.Login(ctx)
			if err != nil {
				return fmt.Errorf("login: %w", err)
			}

			session, err = capExt.GetSession(ctx)
			if err != nil {
				return fmt.Errorf("get session: %w", err)
			}

			log.Info("Logged in with new session")
		} else {
			log.Info("Loaded saved session")
		}

		if session.LoginStatus == extension.LoginStatusSuccess && express.ExpressCheckoutToken != "" {
			break
		}

		log.Info("Login status", "status", session.LoginStatus)
		if session.LoginStatus == extension.LoginStatusChallenge {
			options, err := capExt.OTPGenerate(ctx)
			if err != nil {
				return fmt.Errorf("otp generate: %w", err)
			}

			otp, err := capExt.OTPSend(ctx, options.SmsContactDetails[0].ContactPoint)
			if err != nil {
				return fmt.Errorf("otp send: %w", err)
			}

			pin := ask("Enter SMS OTP")

			_, err = capExt.OTPValidate(ctx, pin, otp.PinAuthenticationToken)
			if err != nil {
				return fmt.Errorf("otp validate: %w", err)
			}

			log.Info("Successfully validated OTP")

			cards, err := capExt.GetPaymentCards(ctx)
			if err != nil {
				return fmt.Errorf("get payment cards: %w", err)
			}

			for _, card := range cards {
				cvv := ask(fmt.Sprintf("Enter CVV for card %s", card.CardNumber))

				err = capExt.ConfigureCard(ctx, card, cvv)
				if err != nil {
					return fmt.Errorf("configure card: %w", err)
				}

				err = capExt.CompleteCardRegister(ctx, card)
				if err != nil {
					return fmt.Errorf("complete card register: %w", err)
				}

				log.Info("Card configured", "card", card.CardNumber)
			}
		}

		if express.ExpressCheckoutToken == "" {
			enrollement, err := capExt.ExpressEnroll(ctx)
			if err != nil {
				return fmt.Errorf("enroll in express: %w", err)
			}

			express.ExpressCheckoutToken = enrollement.ExpressCheckoutToken
		} else {
			login, err := capExt.ExpressLogin(ctx, express.ExpressCheckoutToken)
			if err != nil {
				return fmt.Errorf("login to express: %w", err)
			}

			express.ExpressCheckoutToken = login.ExpressCheckoutToken
		}

		err = profile.Express.Set(express)
		if err != nil {
			return fmt.Errorf("save express: %w", err)
		}
	}

	return nil
}
