export ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(ARGS):;@:) # turn arguments into do-nothing targets
export ARGS

all: empty-page cover eidesstattliche-erklaerung document
.PHONY: all

cover:
	cd $(ROOT_DIR)/00-cover-sheet; \
		latexmk -C; \
		pdflatex --shell-escape index.tex
.PHONY: cover

cut:
	pdftk index.pdf cat $(word 1,$(ARGS))-$(word 2,$(ARGS)) output Zimmermann-MasterThesis-2017-p$(word 1,$(ARGS))-$(word 2,$(ARGS)).pdf
.PHONY: cut

document:
	latexmk -C
	pdflatex --shell-escape index.tex
	bibtex index
	pdflatex --shell-escape index.tex
	pdflatex --shell-escape index.tex
.PHONY: document

eidesstattliche-erklaerung:
	cd $(ROOT_DIR)/00-eidesstattliche-erklaerung; \
		latexmk -C; \
		pdflatex --shell-escape index.tex
.PHONY: eidesstattliche-erklaerung

empty-page:
	cd $(ROOT_DIR)/00-empty-page; \
		latexmk -C; \
		pdflatex --shell-escape index.tex
.PHONY: empty-page

images:
	sh $(ROOT_DIR)/scripts/images-svg-to-pdf.sh
.PHONY: images
