package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/goquizvibe/locales"
)

type contextKey string

const localeKey contextKey = "locale"
const translatorKey contextKey = "translator"

type LocaleMiddleware struct {
	service *locales.Service
}

func NewLocaleMiddleware(service *locales.Service) *LocaleMiddleware {
	return &LocaleMiddleware{service: service}
}

func (m *LocaleMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale := parseAcceptLanguage(r.Header.Get("Accept-Language"))
		translator := m.service.Get(locale)
		ctx := context.WithValue(r.Context(), localeKey, locale)
		ctx = context.WithValue(ctx, translatorKey, translator)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func parseAcceptLanguage(header string) locales.Locale {
	if header == "" {
		return locales.LocaleRu
	}

	parts := strings.SplitSeq(header, ",")
	for part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "en") {
			return locales.LocaleEn
		}
		if strings.HasPrefix(part, "ru") {
			return locales.LocaleRu
		}
	}

	return locales.LocaleRu
}

func GetLocale(ctx context.Context) locales.Locale {
	if locale, ok := ctx.Value(localeKey).(locales.Locale); ok {
		return locale
	}
	return locales.LocaleRu
}

func GetTranslator(ctx context.Context) locales.Translator {
	if t, ok := ctx.Value(translatorKey).(locales.Translator); ok {
		return t
	}
	return nil
}
