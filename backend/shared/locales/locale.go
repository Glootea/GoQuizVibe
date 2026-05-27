package locales

import (
	"context"
)

type contextKey string

const LocaleKey contextKey = "locale"
const TranslatorKey contextKey = "translator"

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

func GetTranslator(ctx context.Context) Translator {
	if t, ok := ctx.Value(TranslatorKey).(Translator); ok {
		return t
	}
	return nil
}
