package locales

import "context"

type Service struct {
	translators map[Locale]Translator
}

func NewService() (*Service, error) {
	s := &Service{
		translators: make(map[Locale]Translator),
	}

	ru, err := New(LocaleRu)
	if err != nil {
		return nil, err
	}
	s.translators[LocaleRu] = ru

	en, err := New(LocaleEn)
	if err != nil {
		return nil, err
	}
	s.translators[LocaleEn] = en

	return s, nil
}

func (s *Service) Get(locale Locale) Translator {
	if t, ok := s.translators[locale]; ok {
		return t
	}
	return s.translators[LocaleRu]
}

func (s *Service) GetForContext(ctx context.Context, locale Locale) Translator {
	return s.Get(locale)
}