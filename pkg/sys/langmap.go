package sys

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type LangMap map[string]interface{}

func NewLangMap(pattern string) (langs LangMap, err error) {
	langs = make(LangMap)
	files, _ := filepath.Glob(pattern)
	for _, f := range files {
		jsonText, err := ioutil.ReadFile(f)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		langCode := strings.Replace(filepath.Base(f), ".json", "", 1)
		langObj := make(map[string]interface{})
		err = json.Unmarshal(jsonText, &langObj)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		langs[langCode] = langObj
	}
	return langs, nil
}

func (lm *LangMap) Locale(lang, text string) string {
	if _, ok := (*lm)[lang]; !ok {
		return ""
	}
	return (*lm)[lang].(map[string]interface{})[text].(string)
}

func (lm *LangMap) GetTemplFunc() (template.FuncMap) {
	return template.FuncMap{"t": lm.Locale}
}