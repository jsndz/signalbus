package types


type TemplateEngine interface {
	Render(data map[string]interface{},tenantID, channel, name, locale string,contentType []string) (string,string,error)
}


