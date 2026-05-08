package pages

import (
	"context"

	"github.com/goquizvibe/middleware"
)

func GetLang(ctx context.Context) string {
	locale := middleware.GetLocale(ctx)
	return string(locale)
}
