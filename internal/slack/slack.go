package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	webhookURL string
}

func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
	}
}

type SlackRequestBody struct {
	Text        string        `json:"text"`
	Parse       string        `json:"parse"`
	Attachments []*Attachment `json:"attachments"`
}

type Attachment struct {
	Color   string `json:"color"`
	Pretext string `json:"pretext"`
	Title   string `json:"title"`
	Text    string `json:"text"`
}

func (c *Client) Send(body *SlackRequestBody) error {
	jsonBody, _ := json.Marshal(*body)

	req, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	client := &http.Client{}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	b := new(bytes.Buffer)
	b.ReadFrom(res.Body)
	if resBody := b.String(); resBody != "ok" {
		return fmt.Errorf("Slack returned not-ok: %q", resBody)
	}

	return nil
}

func (c *Client) SendMessage(text string) error {
	return c.Send(&SlackRequestBody{
		Text:  text,
		Parse: "full",
	})
}
