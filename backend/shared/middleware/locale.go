package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/goquizvibe/backend/shared/locales"
)

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
		ctx := context.WithValue(r.Context(), locales.LocaleKey, locale)
		ctx = context.WithValue(ctx, TranslatorKey, translator)
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

type contextKey string

const TranslatorKey contextKey = "translator"

func GetTranslator(ctx context.Context) locales.Translator {
	if t, ok := ctx.Value(TranslatorKey).(locales.Translator); ok {
		return t
	}
	return nil
}
