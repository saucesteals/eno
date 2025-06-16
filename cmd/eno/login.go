package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/saucesteals/eno/api"
	"github.com/saucesteals/eno/extension"
)

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

			contactPoints := []extension.OPTSmsContactDetails{}
			for _, contactPoint := range options.SmsContactDetails {
				contactPoints = append(contactPoints, contactPoint)
			}

			for _, contactPoint := range options.HomeContactDetails {
				contactPoints = append(contactPoints, contactPoint)
			}

			for _, contactPoint := range options.WorkContactDetails {
				contactPoints = append(contactPoints, contactPoint)
			}

			if len(contactPoints) == 0 {
				return fmt.Errorf("no contact points found")
			}

			fmt.Printf("Contact Points:\n")
			for i, contactPoint := range contactPoints {
				fmt.Printf("%d. %s\n", i+1, contactPoint.ContactPoint)
			}

			contactPointIndex, err := strconv.Atoi(ask("Select a contact point"))
			if err != nil {
				return fmt.Errorf("invalid contact point")
			}

			contactPointIndex--
			if contactPointIndex < 0 || contactPointIndex > len(contactPoints) {
				return fmt.Errorf("unknown contact point")
			}

			otp, err := capExt.OTPSend(ctx, contactPoints[contactPointIndex].ContactPoint)
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
				cvv := ask(fmt.Sprintf("Enter CVV for card %s (%s)", card.CardNumber, card.ProductDescription))

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
