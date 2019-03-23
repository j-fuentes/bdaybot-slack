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

func (c *Client) SendMessage(text string) error {
	body, _ := json.Marshal(slackRequestBody{
		Text:  text,
		Parse: "full",
	})

	req, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewBuffer(body))
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

type slackRequestBody struct {
	Text  string `json:"text"`
	Parse string `json:"parse"`
}
