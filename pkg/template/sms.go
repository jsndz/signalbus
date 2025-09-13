package template

import (
	"bytes"
	"fmt"
	html "html/template"
	text "text/template"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/repositories"
)


func Render(
    data map[string]interface{},
    tenantID, channel, name, locale string,
    contentTypes []string,
    repo *repositories.TemplateRepository,
) (map[string][]byte, error) {
    results := make(map[string][]byte)

    id, err := uuid.Parse(tenantID)
    if err != nil {
        return nil, fmt.Errorf("invalid tenantID: %w", err)
    }

    for _, ct := range contentTypes {
        tmpl, err := repo.GetByLookup(id, channel, name, locale, ct)
        if err != nil {
            return nil, fmt.Errorf("failed to fetch template (%s): %w", ct, err)
        }

        var buf bytes.Buffer

        switch tmpl.ContentType {
        case "text":
            t, err := text.New("tmpl").Parse(tmpl.Content)
            if err != nil {
                return nil, fmt.Errorf("failed to parse text template: %w", err)
            }
            if err := t.Execute(&buf, data); err != nil {
                return nil, fmt.Errorf("failed to render text template: %w", err)
            }

        case "html":
            t, err := html.New("tmpl").Parse(tmpl.Content)
            if err != nil {
                return nil, fmt.Errorf("failed to parse HTML template: %w", err)
            }
            if err := t.Execute(&buf, data); err != nil {
                return nil, fmt.Errorf("failed to render HTML template: %w", err)
            }

        default:
            return nil, fmt.Errorf("unsupported content type: %s", tmpl.ContentType)
        }

        results[tmpl.ContentType] = buf.Bytes()
    }

    return results, nil
}
