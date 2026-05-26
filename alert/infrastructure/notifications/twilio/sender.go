package twilio

import (
    "context"
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "strings"
    "time"

    "microservice-alert-service/alert/infrastructure/configuration"
)

type Sender struct {
    cfg    configuration.Config
    logger *log.Logger
    client *http.Client
}

func NewSender(cfg configuration.Config, logger *log.Logger) *Sender {
    return &Sender{
        cfg:    cfg,
        logger: logger,
        client: &http.Client{Timeout: 10 * time.Second},
    }
}

func (s *Sender) Send(ctx context.Context, to string, body string) error {
    if s.cfg.TwilioAccountSID == "" || s.cfg.TwilioAPIKey == "" || s.cfg.TwilioAPISecret == "" || s.cfg.TwilioPhoneNumber == "" {
        return errors.New("twilio configuration missing")
    }

    endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.cfg.TwilioAccountSID)
    data := url.Values{}
    data.Set("To", to)
    data.Set("From", s.cfg.TwilioPhoneNumber)
    data.Set("Body", body)

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
    if err != nil {
        return err
    }

    req.SetBasicAuth(s.cfg.TwilioAPIKey, s.cfg.TwilioAPISecret)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := s.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("twilio error: %s", string(bodyBytes))
    }

    return nil
}
