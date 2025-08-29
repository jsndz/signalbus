package config

import (
	"fmt"
	"os"

	"github.com/jsndz/signalbus/pkg/gomailer"
	"gopkg.in/yaml.v3"
)


type Config struct{
	Tenants map[string]Tenant `yaml:"tenants"`
}

type Tenant struct{
	Provider string `yaml:"provider"`
	SMTP *gomailer.SMTPMailer
	SendGrid *gomailer.SendGridMailer
}



func UnmarshalYAML()(*map[string]gomailer.Mailer, error){
	var config Config
	data, err := os.ReadFile("config.yml")
	if err!=nil {
		return nil,err
	}
	if err:= yaml.Unmarshal(data,&config);err!=nil{
		return nil,err
	}
	mailerMap := make(map[string]gomailer.Mailer)
	for id,tentantConfig := range config.Tenants{
		switch tentantConfig.Provider {
			case "smtp":{
				mailerMap[id] = &gomailer.SMTPMailer{
					Host:     tentantConfig.SMTP.Host,
					Port:     tentantConfig.SMTP.Port,
					Username: tentantConfig.SMTP.Username,
					Password: tentantConfig.SMTP.Password,
					UseAuth:  tentantConfig.SMTP.UseAuth,
				}
			}
			case "sendgrid":{
				mailerMap[id] = &gomailer.SendGridMailer{
					APIKey:   tentantConfig.SendGrid.APIKey,
					BaseURL:  tentantConfig.SendGrid.BaseURL,
					FromName: tentantConfig.SendGrid.FromName,
					FromMail: tentantConfig.SendGrid.FromMail,
				}
			}
			default:
            	return nil, fmt.Errorf("tenant %s has unsupported provider %s", id, tentantConfig.Provider)
        
		}
	}
	return &mailerMap,nil
}