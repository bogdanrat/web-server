package kvtranslator

import (
	"github.com/bogdanrat/web-server/service/core/i18n"
	"github.com/bogdanrat/web-server/service/core/store"
	"strings"
)

type KeyValueTranslator struct {
	store         store.KeyValue
	keyValuePairs map[string]string
}

func New(keyValueStore store.KeyValue) (i18n.Translator, error) {
	translator := &KeyValueTranslator{
		store:         keyValueStore,
		keyValuePairs: make(map[string]string),
	}

	// TODO: call method when a new key-value pair event was emitted
	if err := translator.Reload(); err != nil {
		return nil, err
	}

	return translator, nil
}

func (t *KeyValueTranslator) Do(key string, substitutions map[string]string) string {
	translation, ok := t.keyValuePairs[key]
	if ok {
		for find, replace := range substitutions {
			translation = strings.Replace(translation, "{{"+find+"}}", replace, -1)
		}
	}
	return translation
}

func (t *KeyValueTranslator) Reload() error {
	keyValuePairs, err := t.store.GetAll()
	if err != nil {
		return err
	}

	t.keyValuePairs = make(map[string]string)
	for _, pair := range keyValuePairs {
		if pairValue, ok := pair.Value.(string); ok {
			t.keyValuePairs[pair.Key] = pairValue
		}
	}

	return nil
}
