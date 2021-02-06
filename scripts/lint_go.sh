#!/bin/bash

GOGC=25 golangci-lint run -c ./config/.golangci.yml
