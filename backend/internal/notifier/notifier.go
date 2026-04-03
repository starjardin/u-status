package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Notifier struct {
	token   string
	from    string
	enabled bool
	client  *http.Client
}

func New(token, from string) *Notifier {
	enabled := token != "" && from != ""
	if !enabled {
		log.Println("notifier: Postmark not configured, email notifications disabled")
	}
	return &Notifier{
		token:   token,
		from:    from,
		enabled: enabled,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *Notifier) SendStatusUp(toEmail, monitorName, monitorURL string) {
	subject := fmt.Sprintf("[u-status] %s is UP", monitorName)
	body := fmt.Sprintf(
		"Good news! Your monitor \"%s\" (%s) is now UP and responding normally.",
		monitorName, monitorURL,
	)
	n.send(toEmail, subject, body)
}

func (n *Notifier) SendStatusDown(toEmail, monitorName, monitorURL, errDetail string) {
	subject := fmt.Sprintf("[u-status] %s is DOWN", monitorName)
	body := fmt.Sprintf(
		"Alert: Your monitor \"%s\" (%s) is DOWN.\n\nError: %s",
		monitorName, monitorURL, errDetail,
	)
	n.send(toEmail, subject, body)
}

type postmarkPayload struct {
	From     string `json:"From"`
	To       string `json:"To"`
	Subject  string `json:"Subject"`
	TextBody string `json:"TextBody"`
}

func (n *Notifier) send(to, subject, body string) {
	if !n.enabled {
		log.Printf("notifier: (disabled) would send to %s: %s", to, subject)
		return
	}

	payload, _ := json.Marshal(postmarkPayload{
		From:     n.from,
		To:       to,
		Subject:  subject,
		TextBody: body,
	})

	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", bytes.NewReader(payload))
	if err != nil {
		log.Printf("notifier: failed to build request: %v", err)
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", n.token)

	resp, err := n.client.Do(req)
	if err != nil {
		log.Printf("notifier: failed to send email to %s: %v", to, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("notifier: sent '%s' to %s", subject, to)
	} else {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("notifier: Postmark error %d for %s: %s", resp.StatusCode, to, respBody)
	}
}
