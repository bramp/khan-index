// tree builds a markdown file of the Khan Academy Index
// by Andrew Brampton (https://bramp.net)
//
// TODO:
// [ ] Some children are "Talkthrough", which have videos, but aren't included yet.
package main

import (
	"encoding/json"
	"flag"
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

var (
	// Count of kinds (for debugging)
	allKinds = make(map[string]int)

	// When displaying, lets rename some of the content kinds.
	kindMapping = map[string]string{
		"Talkthrough":   "Video",
		"TopicQuiz":     "Quiz",
		"TopicUnitTest": "Test",
	}

	// The kinds (in order) for printing
	kinds = []string{"Article", "Challenge", "Exercise", "Interactive", "Project", "Topic", "Quiz", "Test", "Video"}

	// Examples "in-in", "us-cc",  "ca-on", "ca-ab"
	// "in-in" India
	// "us-cc" US - Common Core standards
	// "ca-on" Canada - Ontario
	// "ca-ab" Canada - Alberta
	curriculumKey   = flag.String("curriculum_key", "", "regex filter to this curriculum_key")
	curriculumKeyRe *regexp.Regexp

	printToc = flag.Bool("toc", false, "print table of content")
)

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

	CurriculumKey string `json:"curriculum_key"`

	Slug            string
	DomainSlug      string `json:"domain_slug"`
	Title           string `json:"translated_title"`
	StandaloneTitle string `json:"translated_standalone_title"`
	Description     string `json:"translated_description"`

	CreationDate string `json:"creation_date"` // e.g "2018-09-26T09:40:13Z"
	RenderType   string `json:"render_type"`   // e.g "Root", "Domain", "Subject", "Topic", "Tutorial"

	Url         string `json:"ka_url"`
	UserLicense string `json:"ka_user_license"` // e.g "cc-by-nc-sa"

	ListedLocales []string `json:"listed_locales"` // Seems inconsisently set

	Deleted bool

	YoutubeId string `json:"translated_youtube_id"`
	Duration  int

	Children  []Topic
	ChildData []Child `json:"child_data"`
}

// Returns true iff this node (and sub nodes) should be excluded out.
func (t *Topic) Exclude() bool {
	// Only descend into matching curricula
	return t.RenderType == "Subject" && !curriculumKeyRe.Match([]byte(t.CurriculumKey))
}

// Returns the number children (which are not excluded).
func (t *Topic) ChildCount() int {
	if t.RenderType != "Domain" {
		return len(t.Children)
	}

	// Domains may have children we exclude
	count := 0
	for _, child := range t.Children {
		if !child.Exclude() {
			count++
		}
	}

	return count
}

func main() {
	flag.Parse()

	curriculumKeyRe = regexp.MustCompile(*curriculumKey)

	if flag.NArg() != 3 {
		fmt.Fprintf(os.Stderr, "Usage go run tree.go [topictree.json] [index.md] [index.yaml]\n")
		return
	}

	// Fetch from curl http://www.khanacademy.org/api/v1/topictree
	f, err := os.Open(flag.Arg(0))
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

	markdown, err := os.Create(flag.Arg(1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create markdown: %s\n", err)
		return
	}
	defer markdown.Close()

	if *printToc {
		toc(markdown, &root)
	}

	dfs(markdown, &root, 0)

	metadata, err := os.Create(flag.Arg(2))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create metadata: %s\n", err)
		return
	}
	defer metadata.Close()

	meta(metadata, &root)
}

func parseLanguageFromFilename(filename string) string {
	re := regexp.MustCompile("(..)\\..*")
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
	fmt.Fprintf(out, "url: %s\n", root.Url)
	fmt.Fprintf(out, "lang: %s\n", parseLanguageFromFilename(flag.Arg(2)))
	fmt.Fprintf(out, "---\n")
}

func toc(out io.Writer, t *Topic) {
	// Only descend into matching curricula
	if t.Exclude() {
		return
	}

	if t.ChildCount() == 0 {
		return
	}

	if t.RenderType == "Domain" {
		fmt.Fprintf(out, "\n# %s\n", t.StandaloneTitle)
	} else if t.RenderType == "Subject" {
		fmt.Fprintf(out, "## %s\n", t.StandaloneTitle)
	} else if t.RenderType == "Root" {
		// Do nothing
	} else {
		// All other types we should just bail
		return
	}

	for _, child := range t.Children {
		toc(out, &child)
	}
}

func dfs(out io.Writer, t *Topic, depth int) {

	if t.ChildCount() == 0 {
		return
	}

	// Only descend into matching curricula
	if t.Exclude() {
		return
	}

	print(out, t, depth)

	for _, child := range t.Children {
		dfs(out, &child, depth+1)
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

	fmt.Fprintf(out, "[%s](%s)", t.StandaloneTitle, t.Url)

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
		q.Set("title", t.StandaloneTitle)

		url := "https://www.youtube.com/watch_videos?" + q.Encode()

		// BUG: The duration does not include the duration of any Talkthrough videos.
		fmt.Fprintf(out, " (%s | [%s](%s) )\n", kindsString(kinds),
			durationString(time.Duration(duration)*time.Second), url)
	} else {
		fmt.Fprintf(out, " (%s)\n", kindsString(kinds))
	}
}
