package adapters

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type SMSAdapter struct {
	TwillioSID      string
	TwillioPassword string
	TwillioNumber   string
}

func NewSMSAdapter() *SMSAdapter {
	return &SMSAdapter{
		TwillioSID:      os.Getenv("TWILLIO_ACCOUNT_SID"),
		TwillioPassword: os.Getenv("TWILLIO_PASSWORD"),
		TwillioNumber:   os.Getenv("TWILLIO_NUMBER"),
	}
}

func (s *SMSAdapter) Send(to, message string) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.TwillioSID,
		Password: s.TwillioPassword,
	})

	params := &api.CreateMessageParams{}
	params.SetBody(message)
	params.SetFrom(s.TwillioNumber)
	params.SetTo(to)

	_, err := client.Api.CreateMessage(params)
	if err != nil {
		return err
	}

	fmt.Println("📨 SMS sent successfully")
	return nil
}
