build:
    go build .
deploy: build
    gh extension remove gh-template
    gh extension install .
