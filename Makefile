.PHONY: clean all docs
.PRECIOUS: data/topictree.%.json docs/%.md

all: docs/index.html docs

docs: docs/bg.html docs/bn.html docs/cs.html docs/da.html docs/de.html 
docs: docs/en.html docs/es.html docs/fr.html docs/gu.html docs/hi.html 
docs: docs/hy.html docs/id.html docs/it.html docs/ja.html docs/ka.html 
docs: docs/ko.html docs/mn.html docs/nb.html docs/nl.html docs/pl.html 
docs: docs/pt-pt.html docs/pt.html docs/sr.html docs/sv.html docs/ta.html 
docs: docs/tr.html docs/zh-hans.html
docs: docs/in.html docs/ca.html

clean:
	rm docs/*

docs/index.md: docs
	go run index.go > $@

docs/index.html: docs/index.md header.md footer.md
	pandoc \
		-f gfm -t html -s                          \
		-M title="(Unofficial) Khan Academy Index" \
		-V url="https://khanacademy.org/" \
		--lua-filter meta-vars.lua    \
		--lua-filter=change-links.lua \
		header.md $< footer.md -o $@

docs/%.html: docs/%.md header.md footer.md
	pandoc \
		-f gfm -t html -s --toc       \
		--metadata-file=docs/$*.yaml  \
		--lua-filter meta-vars.lua    \
		header.md $< footer.md -o $@

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
