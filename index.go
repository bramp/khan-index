// Index scan all docs/*.md and write out a simple index.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var (
	// The list of languages
	languages = map[string]string{
		"az":      "Azərbaycanca",
		"bg":      "български",
		"bn":      "বাংলা", // Bengali
		"ca":      "Canadian",
		"cs":      "čeština",
		"da":      "dansk",
		"de":      "Deutsch",
		"en":      "English",
		"es":      "español",
		"fr":      "français",
		"gu":      "ગુજરાતી",
		"hi":      "हिन्दी", // Hindi
		"hu":      "magyar",
		"hy":      "հայերեն", // Armenian
		"id":      "Bahasa Indonesia",
		"it":      "italiano",
		"in":      "Indian",
		"ja":      "日本語", // Japanese
		"ka":      "ქართული",
		"km":      "ខ្មែរ",
		"kn":      "ಕನ್ನಡ",
		"ko":      "한국어",
		"mn":      "монгол",
		"my":      "ဗမာ",
		"nb":      "norsk bokmål",
		"nl":      "Nederlands",
		"pl":      "polski",
		"pt":      "português (Brazil)",
		"pt-pt":   "português (Portugal)",
		"ru":      "русский",
		"sr":      "Српски",
		"sv":      "svenska",
		"ta":      "தமிழ்",
		"tr":      "Türkçe",
		"uz":      "Oʻzbek",
		"zh-hans": "中文 (简体中文)",
	}
)

func main() {
	matches, err := filepath.Glob("docs/*.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read docs: %s", err)
	}

	re := regexp.MustCompile(`([^./]+)\..+`)

	// TODO Sort matches

	for _, filename := range matches {
		match := re.FindStringSubmatch(filename)
		if len(match) < 1 {
			fmt.Fprintf(os.Stderr, "Cannot extract language from filename %q", filename)
			continue
		}
		lang := match[1]
		if lang == "index" {
			continue
		}

		language, ok := languages[lang]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown language %q", lang)
			continue
		}

		// Perhaps get something that maps language codes to country codes
		flagRemap := map[string]string{
			"da":      "dk",
			"en":      "gb",
			"hi":      "in", // Hindi (Indian)
			"hy":      "am",
			"nb":      "no",
			"pt":      "br",
			"pt-pt":   "pt",
			"ko":      "kr",
			"ja":      "jp",
			"ka":      "ge", // Georgia
			"ta":      "lk", //  Tamil people of India and Sri Lanka
			"zh-hans": "cn",
		}
		flag := flagRemap[lang]
		if flag == "" {
			flag = lang
		}

		path := filepath.Base(filename)
		fmt.Printf("* ![%s flag](flags/png/%s.png) [%s](%s)\n", language, flag, language, path)
	}
}
