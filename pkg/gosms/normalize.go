package gosms

import (
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

func NormalizeSMS(num string) (string, error) {
    if num == "" {
        return "", fmt.Errorf("missing number")
    }
    if num[0] != '+' {
        return "", fmt.Errorf("phone number must be in E.164 format with +")
    }

    parsed, err := phonenumbers.Parse(num, "")
    if err != nil || !phonenumbers.IsValidNumber(parsed) {
        return "", fmt.Errorf("invalid phone number")
    }

    return phonenumbers.Format(parsed, phonenumbers.E164), nil
}
