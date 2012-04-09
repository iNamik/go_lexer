GOROOT ?= $(shell printf 't:;@echo $$(GOROOT)\n' | gomake -f -)
include $(GOROOT)/src/Make.inc

TARG=github.com/iNamik/go_lexer

GOFILES=\
	impl.go\
	lexer.go\
	private.go\

include $(GOROOT)/src/Make.pkg

