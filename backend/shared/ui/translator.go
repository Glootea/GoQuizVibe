package ui

import (
	"context"

	"github.com/goquizvibe/backend/shared/middleware"
)

func GetLang(ctx context.Context) string {
	locale := middleware.GetLocale(ctx)
	return string(locale)
}