package locales

import (
	"context"
)

const LocaleKey string = "locale"

func GetLang(ctx context.Context) string {
	locale := GetLocale(ctx)
	return string(locale)
}

func GetLocale(ctx context.Context) Locale {
	if locale, ok := ctx.Value(LocaleKey).(Locale); ok {
		return locale
	}
	return LocaleRu
}
