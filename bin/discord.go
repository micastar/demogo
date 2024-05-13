package bin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Discord struct {
	Content string `json:"content"`
}

type Config struct {
	WebhookURL string `json:"webhook"`
}

func (c *Config) NewDiscord(msg Post) Discord {

	discord := Discord{
		Content: fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\t%s", msg.Title, msg.Descrp, msg.Link, msg.Cates, msg.Logdate),
	}

	return discord
}

func (c *Config) InitialConfig(msg Post) error {
	discord := c.NewDiscord(msg)

	err := c.SendDiscordReq(discord)
	if err != nil {
		return fmt.Errorf("InitialConfig Error: %s", err)
	}
	return nil

}

func (c *Config) SendDiscordReq(discord Discord) error {
	json, err := json.Marshal(discord)
	if err != nil {
		log.Printf("%s", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("Error response from Discord with request body <%s> [%d] - [%s]", json, resp.StatusCode, string(buf))
	}
	return nil
}
