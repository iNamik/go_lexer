GOROOT ?= $(shell printf 't:;@echo $$(GOROOT)\n' | gomake -f -)
include $(GOROOT)/src/Make.inc

TARG=word_count

GOFILES=\
	word_count.go\

include $(GOROOT)/src/Make.cmd

