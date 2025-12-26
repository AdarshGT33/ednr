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
	client := twilio.NewRestClient()

	params := &api.CreateMessageParams{}
	params.SetBody(message)
	params.SetFrom(s.TwillioNumber)
	params.SetTo(to)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		if resp.Body != nil {
			fmt.Println(*resp.Body)
		} else {
			fmt.Println(*resp.Body)
		}
	}
	return nil
}
