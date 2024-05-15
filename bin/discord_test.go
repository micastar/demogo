package bin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/micastar/discord-feed/global"
)

func TestDiscordWebhook(t *testing.T) {
	t.Run("sendWebhook", func(t *testing.T) {
		err := godotenv.Load("../.env")
		if err != nil {
			t.Errorf("Error loading .env file: %v", err)
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		client := &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: global.DisableKeepAlives,
			},
			Timeout: global.DefaultTimeOut,
		}

		c := &Config{}
		c.WebhookURL = os.Getenv("WebhookURL")

		err = c.SendDiscordReq(c.NewDiscord(&Post{}), client)

		if err != nil {
			t.Errorf("sendDiscordReq failed: %v", err)
		}
	})

}
