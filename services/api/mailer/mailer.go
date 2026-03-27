package mailer

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v3"
)

type Service struct {
	client *resend.Client
	from   string
}

func New() (*Service, error) {
	key := os.Getenv("RESEND_API_KEY")
	from := os.Getenv("RESEND_FROM")

	if key == "" {
		return nil, fmt.Errorf("missing RESEND_API_KEY")
	}
	if from == "" {
		return nil, fmt.Errorf("missing RESEND_FROM")
	}

	return &Service{
		client: resend.NewClient(key),
		from:   from,
	}, nil
}

func (s *Service) SendOrderConfirmation(ctx context.Context, to string, orderID string) (string, error) {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: "Order confirmation",
		Html: fmt.Sprintf(`
<h1>Thanks for your order</h1>

<p>Hi Tim,</p>

<p>We’ve received your payment and your order is now being processed.</p>

<p><strong>Order ID:</strong>%s</p>

<p><strong>Shipping address:</strong><br>
Nyholms Alle 6A, 2TV<br>
2610 Rødovre<br>
Danmark
</p>

<p>If you have any questions, reply to this email.</p>

<p>— Copenhagen Collectibles</p>
		`, orderID),
		ReplyTo: "support@yourdomain.com",
	}

	resp, err := s.client.Emails.Send(params)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("resend returned nil response")
	}

	return resp.Id, nil
}
