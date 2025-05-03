SHELL := /bin/bash
SHELLFLAGS := -c
.ONESHELL:
CURR_PATH := $(shell pwd)

install:
	@ go env -w GOBIN=$(CURR_PATH)/bin
	@ go install




.PHONY: install