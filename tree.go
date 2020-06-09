// tree builds a markdown file of the Khan Academy Index
// by Andrew Brampton (https://bramp.net)
//
// TODO:
// [ ] Some children are "Talkthrough", which have videos, but aren't included yet.
package main

/*





   1                     "curriculum_key": "ca-ab",
   1                     "curriculum_key": "ca-on",
   1             "curriculum_key": "in-in",
   1             "curriculum_key": "us-cc",
   1     "curriculum_key": "",
  12             "curriculum_key": "",
  14                     "curriculum_key": null,
  29                     "curriculum_key": "us-cc",
  38                     "curriculum_key": "in-in",
*/

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
)

var languages = map[string]string{
	"az":      "Azərbaycanca",
	"bg":      "български",
	"bn":      "বাংলা", // Bengali
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

// The kinds (in order) for printing
var kinds = []string{"Article", "Challenge", "Exercise", "Interactive", "Project", "Topic", "Quiz", "Test", "Video"}

// For displaying, lets rename some of the content kinds.
var kindMapping = map[string]string{
	"Talkthrough":   "Video",
	"TopicQuiz":     "Quiz",
	"TopicUnitTest": "Test",
}

type Child struct {
	Id   string
	Kind string
}

type Topic struct {
	Id string

	// Kind is the kind of topic. Count of by Kind:
	//	Article:8257
	//	Challenge:166
	//	Exercise:15754
	//	Interactive:80
	//	Project:36
	//	Talkthrough:131     # Has videos
	//	Topic:11243
	//	TopicQuiz:2092
	//	TopicUnitTest:920
	//	Video:31448         # Is a video
	Kind string

	Slug        string
	DomainSlug  string `json:"domain_slug"`
	Title       string `json:"translated_title"`
	Description string `json:"translated_description"`

	CreationDate string `json:"creation_date"` // e.g "2018-09-26T09:40:13Z"
	RenderType   string `json:"render_type"`   // e.g "Root", "Domain", "Subject", "Topic", "Tutorial"

	Url         string `json:"ka_url"`
	UserLicense string `json:"ka_user_license"` // e.g "cc-by-nc-sa"

	ListedLocales []string `json:"listed_locales"` // Seems inconsisently set

	Deleted bool

	YoutubeId string `json:"youtube_id"`
	Duration  int

	Children  []Topic
	ChildData []Child `json:"child_data"`
}

var (
	allKinds = make(map[string]int)
)

func main() {

	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage go run tree.go [topictree.json] [index.md] [index.yaml]\n")
		return
	}

	// Fetch from curl http://www.khanacademy.org/api/v1/topictree
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open topic tree: %s\n", err)
		return
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	var root Topic
	if err := dec.Decode(&root); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode topictree: %s\n", err)
		return
	}

	markdown, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create markdown: %s\n", err)
		return
	}
	defer markdown.Close()

	//toc(markdown, &root, 0)
	dfs(markdown, &root, 0)

	metadata, err := os.Create(os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create metadata: %s\n", err)
		return
	}
	defer metadata.Close()

	meta(metadata, &root)

	//fmt.Fprintf(os.Stderr, "Kinds: %v\n", allKinds)
}

func parseLanguageFromFilename(filename string) string {
	re := regexp.MustCompile("\\.(.+)\\.json")
	match := re.FindStringSubmatch(filename)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func durationString(d time.Duration) string {
	s := ""
	if d >= time.Hour {
		s += strconv.Itoa(int(d/time.Hour)) + "h "
		d = d % time.Hour
	}
	if d >= time.Minute {
		s += strconv.Itoa(int(d/time.Minute)) + "m "
		d = d % time.Minute
	}
	if d >= time.Second {
		s += strconv.Itoa(int(d/time.Second)) + "s "
		d = d % time.Second
	}
	return strings.TrimSpace(s)
}

func countChildKinds(children []Child) map[string]int {
	m := make(map[string]int)

	for _, child := range children {
		m[child.Kind]++
	}

	return m
}

func kindsString(m map[string]int) string {

	// Map to nicer names
	display := make(map[string]int)
	for k, v := range m {
		if mapping, found := kindMapping[k]; found {
			k = mapping
		}
		display[k] += v
	}

	var s strings.Builder
	for _, kind := range kinds {
		if display[kind] > 0 {
			fmt.Fprintf(&s, "%s, ", english.Plural(display[kind], kind, ""))
		}
	}
	return strings.TrimSuffix(s.String(), ", ")
}

func meta(out io.Writer, root *Topic) {
	fmt.Fprintf(out, "---\n")
	fmt.Fprintf(out, "title: %s\n", root.Title)
	fmt.Fprintf(out, "author: %s\n", "Khan Academy")
	fmt.Fprintf(out, "lang: %s\n", parseLanguageFromFilename(os.Args[1]))
	fmt.Fprintf(out, "---\n")
}

func dfs(out io.Writer, t *Topic, depth int) {
	if len(t.ChildData) == 0 {
		return
	}

	print(out, t, depth)

	for _, child := range t.Children {
		dfs(out, &child, depth+1)
	}
}

func toc(out io.Writer, t *Topic, depth int) {
	if len(t.ChildData) == 0 {
		return
	}

	if t.RenderType == "Domain" {
		fmt.Fprintf(out, "\n# %s\n", t.Title)
	} else if t.RenderType == "Subject" {
		fmt.Fprintf(out, "## %s\n", t.Title)
	}

	for _, child := range t.Children {
		toc(out, &child, depth+1)
	}
}

func print(out io.Writer, t *Topic, depth int) {
	if t.RenderType == "Root" {
		// Do nothing for the root
		return
	}
	if t.RenderType == "Domain" {
		fmt.Fprintf(out, "\n\n# ")
	} else if t.RenderType == "Subject" {
		fmt.Fprintf(out, "\n## ")
	} else {
		// Minus the domain and the subject indent
		listDepth := depth - 2
		if listDepth < 0 {
			listDepth = 0
		}

		indent := strings.Repeat("  ", listDepth)
		fmt.Fprintf(out, "%s- ", indent)
	}

	fmt.Fprintf(out, "[%s](%s)", t.Title, t.Url)

	kinds := countChildKinds(t.ChildData)
	for k, v := range kinds {
		allKinds[k] += v
	}

	var youtubeIds []string
	duration := 0
	for _, child := range t.Children {
		duration += child.Duration
		youtubeIds = append(youtubeIds, child.YoutubeId)
	}

	if duration > 0 {
		q := url.Values{}
		q.Set("video_ids", strings.Join(youtubeIds, ","))
		q.Set("title", t.Title)

		url := "https://www.youtube.com/watch_videos?" + q.Encode()

		// BUG: The duration does not include the duration of any Talkthrough videos.
		fmt.Fprintf(out, " (%s | [%s](%s) )\n", kindsString(kinds),
			durationString(time.Duration(duration)*time.Second), url)
	} else {
		fmt.Fprintf(out, " (%s)\n", kindsString(kinds))
	}
}
