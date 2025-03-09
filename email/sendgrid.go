package email

import (
    "errors"
    "fmt"
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
    dbThings "ibooks_notes_exporter/db"
    "os"
)

// Config holds email configuration
type Config struct {
    APIKey      string
    FromEmail   string
    FromName    string
    ToEmail     string
    ToName      string
    Subject     string
}

// SendRandomNote sends a random note via email
func SendRandomNote(config Config, note dbThings.RandomNote) error {
    if config.APIKey == "" {
        return errors.New("Sendgrid API key is required")
    }

    from := mail.NewEmail(config.FromName, config.FromEmail)
    to := mail.NewEmail(config.ToName, config.ToEmail)
    
    // Create email content
    htmlContent := fmt.Sprintf(`
        <h1>%s — %s</h1>
        <blockquote>%s</blockquote>
        %s
    `, note.BookTitle, note.BookAuthor, note.Highlight, formatNote(note.Note))

    message := mail.NewSingleEmail(from, config.Subject, to, "", htmlContent)
    client := sendgrid.NewSendClient(config.APIKey)
    response, err := client.Send(message)
    
    if err != nil {
        return err
    }
    
    if response.StatusCode >= 400 {
        return fmt.Errorf("error sending email: %d - %s", response.StatusCode, response.Body)
    }
    
    return nil
}

// GetConfigFromEnv gets email configuration from environment variables
func GetConfigFromEnv() Config {
    return Config{
        APIKey:      os.Getenv("SENDGRID_API_KEY"),
        FromEmail:   getEnvWithDefault("EMAIL_FROM_ADDRESS", "notifications@interweb.observer"),
        FromName:    getEnvWithDefault("EMAIL_FROM_NAME", "Notes"),
        ToEmail:     getEnvWithDefault("EMAIL_TO_ADDRESS", "cooper@vanwijck.me"),
        ToName:      getEnvWithDefault("EMAIL_TO_NAME", "Reader"),
        Subject:     getEnvWithDefault("EMAIL_SUBJECT", "Your Random Book Note"),
    }
}

func getEnvWithDefault(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func formatNote(note string) string {
    if note == "" {
        return ""
    }
    return fmt.Sprintf("<p>%s</p>", note)
}