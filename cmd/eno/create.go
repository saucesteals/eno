package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/saucesteals/eno/api"
	"github.com/saucesteals/eno/extension"
	"github.com/saucesteals/eno/web"
)

type CreateMode string

var (
	CreateModeWeb       CreateMode = "web"
	CreateModeExtension CreateMode = "extension"
)

func create(ctx context.Context, profile *Profile, capWeb *web.Web, capExt *extension.Extension, card extension.PaymentCard) error {
	mode := CreateMode(ask(fmt.Sprintf("Enter mode (%s/%s)", CreateModeWeb, CreateModeExtension)))
	if mode != CreateModeWeb && mode != CreateModeExtension {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	count, err := strconv.Atoi(ask("Enter number of cards to create"))
	if err != nil {
		return fmt.Errorf("invalid number of cards: %w", err)
	}

	var merchant *extension.DataSource
	var cardPrefix string
	if mode == CreateModeWeb {
		cardPrefix = "Web"

		assessment, err := capWeb.ChallengeAssessment(ctx, card)
		if err != nil {
			return fmt.Errorf("challenge assessment: %w", err)
		}

		if assessment.RedirectURL == "" {
			if len(assessment.AvailableMethods) == 0 {
				return fmt.Errorf("no available methods found")
			}

			smsContactPoints := []web.ChallengeContactPoint{}
			for _, contactPoint := range assessment.AvailableMethods[0].AvailableMethodsPayload.ContactPoints {
				if contactPoint.ContactPointDeliveryMediums.IsSms {
					smsContactPoints = append(smsContactPoints, contactPoint)
				}
			}

			if len(smsContactPoints) == 0 {
				return fmt.Errorf("no sms contact points found")
			}

			fmt.Printf("Contact Points:\n")
			for i, contactPoint := range smsContactPoints {
				fmt.Printf("%d. %s\n", i+1, contactPoint.ContactPointMasked)
			}

			contactPointIndex, err := strconv.Atoi(ask("Select a contact point"))
			if err != nil {
				return fmt.Errorf("invalid contact point")
			}

			contactPointIndex--
			if contactPointIndex < 0 || contactPointIndex > len(smsContactPoints) {
				return fmt.Errorf("unknown contact point")
			}

			smsContactPoint := smsContactPoints[contactPointIndex]

			otp, err := capWeb.ChallengeVerification(ctx, assessment.PolicyProcessID, smsContactPoint)
			if err != nil {
				return fmt.Errorf("challenge verification: %w", err)
			}

			otpValue := ask("Enter OTP value sent to " + smsContactPoint.ContactPointMasked)
			if err := capWeb.ChallengeValidation(ctx, assessment.PolicyProcessID, otp.Otp, otpValue); err != nil {
				return fmt.Errorf("challenge validation: %w", err)
			}
		}
	} else {
		merchantUrl := ask("Enter merchant URL (e.g. www.google.com)")

		m, err := capExt.DataSourceSearch(ctx, merchantUrl)
		if err != nil {
			return fmt.Errorf("failed to search for merchant: %w", err)
		}

		merchant = &m
		cardPrefix = merchant.Name
	}

	w, err := NewCardWriter(profile, card, cardPrefix)
	if err != nil {
		return fmt.Errorf("new card writer: %w", err)
	}
	defer w.Close()

	delay := time.Second * 5
	maxTries := 3
	for i := range count {
		name := fmt.Sprintf("%s Card %d", cardPrefix, i+1)
		var token api.Token
		for j := range maxTries {
			if mode == CreateModeWeb {
				token, err = capWeb.CreateToken(ctx, name, card)
			} else {
				token, err = capExt.CreateToken(ctx, name, card, *merchant)
			}
			if err != nil {
				if j == maxTries-1 {
					return fmt.Errorf("create token: %w", err)
				}
				log.Error("Failed to create token", "error", err)

				if errors.Is(err, api.ErrRateLimited) {
					log.Info("Rate limited, sleeping for 2 minutes")
					time.Sleep(time.Minute * 2)
					continue
				}

				continue
			}

			break
		}

		log.Info(fmt.Sprintf("(%d/%d) Created token", i+1, count), "token", token.Token)

		err = w.Write(token)
		if err != nil {
			return fmt.Errorf("write token: %w", err)
		}

		if i < count-1 {
			time.Sleep(delay)
		}
	}

	log.Info("Created cards", "count", count, "path", w.GetPath())
	return nil
}
