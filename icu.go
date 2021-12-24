package i18n

import (
	"regexp"
	"strings"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	icuLeftDelimiter  = '{'
	icuRightDelimiter = '}'
)

var (
	icuRegexp  = regexp.MustCompile(`{\s*(\w*),\s*plural,`)
	pluralKeys = map[string]string{
		"0":     "zero",
		"zero":  "zero",
		"1":     "one",
		"one":   "one",
		"2":     "two",
		"two":   "two",
		"few":   "few",
		"many":  "many",
		"other": "other",
	}
)

// ParseICUString decodes the text string with ICU format which only supports
// the plural syntax on starling platform.
func ParseICU(text string) (varName string, msg *goi18n.Message, err error) {
	idxArr := icuRegexp.FindStringIndex(text)
	subs := icuRegexp.FindStringSubmatch(text)
	if len(idxArr) != 2 || len(subs) != 2 {
		err = ErrInvalidICUFormat
		return
	}
	varName = subs[1]
	start, end, count := idxArr[0], 0, 0
	for end = start; end < len(text); end++ {
		if text[end] == icuLeftDelimiter {
			count++
		} else if text[end] == icuRightDelimiter {
			count--
		}
		if count == 0 {
			break
		}
	}
	if start < 0 || end < 0 || end > len(text)-1 {
		err = ErrInvalidICUFormat
		return
	}

	pluralMap := make(map[string]string)
	for key := range pluralKeys {
		k := key + " " + string(icuLeftDelimiter)
		if !strings.Contains(text, k) {
			continue
		}
		s := strings.Index(text, k)
		e, cnt := 0, 1
		for e = s + len(k); e < len(text); e++ {
			if text[e] == icuLeftDelimiter {
				cnt++
			}
			if text[e] == icuRightDelimiter {
				cnt--
			}
			if cnt == 0 {
				break
			}
		}
		pluralMap[key] = text[s+len(k) : e]
	}
	preSentence := text[0:start]
	postSentence := text[end+1:]
	msg = &goi18n.Message{}
	for key, value := range pluralMap {
		sentence := preSentence + value + postSentence
		if err = buildExtra(msg, key, sentence); err != nil {
			return
		}
	}
	return
}

func buildExtra(textExtra *goi18n.Message, key, value string) error {
	switch key {
	case "0", "zero":
		textExtra.Zero = value
	case "1", "one":
		textExtra.One = value
	case "2", "two":
		textExtra.Two = value
	case "few":
		textExtra.Few = value
	case "many":
		textExtra.Many = value
	case "other":
		textExtra.Other = value
	default:
		return ErrInvalidICUFormat
	}
	return nil
}
