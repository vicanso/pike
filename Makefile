.PHONY: default web build

web:
	flutter run -d chrome --web-hostname=127.0.0.1 --web-port=3123

lint:
	dartanalyzer --options analysis_options.yaml .
format:
	flutter format .
build:
	flutter build web