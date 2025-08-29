package config

import (
	"fmt"
	"os"

	"github.com/jsndz/signalbus/pkg/gomailer"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Email ServiceConfig `yaml:"email"`
}

type ServiceConfig struct {
	Provider string          `yaml:"provider"`
	SMTP     *gomailer.SMTPMailer     `yaml:"smtp,omitempty"`
	SendGrid *gomailer.SendGridMailer `yaml:"sendgrid,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func BuildMailer(cfg *Config) (gomailer.Mailer, error) {
	switch cfg.Email.Provider {
	case "smtp":
		if cfg.Email.SMTP == nil {
			return nil, fmt.Errorf("missing smtp config for email provider")
		}
		return &gomailer.SMTPMailer{
			Host:     cfg.Email.SMTP.Host,
			Port:     cfg.Email.SMTP.Port,
			Username: cfg.Email.SMTP.Username,
			Password: cfg.Email.SMTP.Password,
			UseAuth:  cfg.Email.SMTP.UseAuth,
		}, nil

	case "sendgrid":
		if cfg.Email.SendGrid == nil {
			return nil, fmt.Errorf("missing sendgrid config for email provider")
		}
		return &gomailer.SendGridMailer{
			APIKey:   cfg.Email.SendGrid.APIKey,
			BaseURL:  cfg.Email.SendGrid.BaseURL,
			FromName: cfg.Email.SendGrid.FromName,
			FromMail: cfg.Email.SendGrid.FromMail,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.Email.Provider)
	}
}
