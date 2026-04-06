package email

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

type Client struct {
	client      *resend.Client
	fromAddress string
	enabled     bool
}

func New() *Client {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromAddress := os.Getenv("RESEND_FROM_ADDRESS")
	if fromAddress == "" {
		fromAddress = "StatusPage <notifications@statuspage.dev>"
	}

	if apiKey == "" {
		log.Println("RESEND_API_KEY not set, email notifications disabled")
		return &Client{enabled: false}
	}

	return &Client{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		enabled:     true,
	}
}

func (c *Client) Enabled() bool {
	return c.enabled
}

func (c *Client) SendIncidentNotification(ctx context.Context, to []string, subject, htmlBody string) error {
	if !c.enabled {
		return nil
	}
	if len(to) == 0 {
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    c.fromAddress,
		To:      to,
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := c.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
