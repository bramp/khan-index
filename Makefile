.PHONY: clean
.PRECIOUS: data/topictree.%.json docs/%.md

all: docs/bg.html docs/bn.html docs/cs.html docs/da.html docs/de.html 
all: docs/en.html docs/es.html docs/fr.html docs/gu.html docs/hi.html 
all: docs/hy.html docs/id.html docs/it.html docs/ja.html docs/ka.html 
all: docs/ko.html docs/mn.html docs/nb.html docs/nl.html docs/pl.html 
all: docs/pt-pt.html docs/pt.html docs/sr.html docs/sv.html docs/ta.html 
all: docs/tr.html docs/zh-hans.html
all: docs/in.html docs/ca.html

#docs/index.md: index.go
#	go run index.go

docs/%.html: docs/%.md header.md footer.md
	pandoc --lua-filter meta-vars.lua \
		-f gfm -t html -s --metadata-file=docs/$*.yaml \
		header.md $< footer.md -o $@
#	--template pandoc-uikit/standalone.html --toc --toc-depth=2

docs/%.md: data/topictree.%.json tree.go 
	go run tree.go $< docs/$*.md docs/$*.yaml

# The topictree.en.json has a few different languages, lets manually break them out.
docs/ca.md: data/topictree.en.json tree.go
	go run tree.go --curriculum_key='ca-.*' $< docs/ca.md docs/ca.yaml

docs/en.md: data/topictree.en.json tree.go
	go run tree.go --curriculum_key='^(us-cc|)$$' $< docs/en.md docs/en.yaml

docs/in.md: data/topictree.en.json tree.go
	go run tree.go --curriculum_key='in-in' $< docs/in.md docs/in.yaml

data/topictree.%.json:
	mkdir -p data
	curl https://$*.khanacademy.org/api/v1/topictree | python -m json.tool > $@

clean:
	rm docs/*