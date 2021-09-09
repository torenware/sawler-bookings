package render

import "github.com/tsawler/bookings-app/internal/models"

func BuildMailHTML(template string, fields map[string]string, data models.MailData) (string, error) {
	fields["To"] = data.To
	fields["From"] = data.From
	fields["Subject"] = data.Subject
	tdata := models.TemplateData{
		StringMap: fields,
	}

	content, err := TemplateAsString(template, &tdata )
	if err != nil {
		return "", err
	}
	return content, nil
}

