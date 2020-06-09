.PHONY: clean
.PRECIOUS: data/topictree.%.json docs/%.md

all: docs/bg.html docs/bn.html docs/cs.html docs/da.html docs/de.html docs/en.html docs/es.html docs/fr.html docs/gu.html docs/hi.html docs/hy.html docs/id.html docs/it.html docs/ja.html docs/ka.html docs/ko.html docs/mn.html docs/nb.html docs/nl.html docs/pl.html docs/pt-pt.html docs/pt.html docs/sr.html docs/sv.html docs/ta.html docs/tr.html docs/zh-hans.html

docs/%.html: docs/%.md
	pandoc $< -f gfm -t html -s --metadata-file=docs/$*.yaml -o $@ 
#	--template pandoc-uikit/standalone.html --toc --toc-depth=2

docs/%.md: data/topictree.%.json tree.go 
	go run tree.go $< docs/$*.md docs/$*.yaml

data/topictree.%.json:
	curl https://$*.khanacademy.org/api/v1/topictree | python -m json.tool > $@

clean:
	rm docs/*