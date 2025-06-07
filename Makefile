.PHONY: all clean docs completions

all: box

box: box.go go.mod go.sum
	go build -o box

build: box

clean:
	rm -f box man/box.1
	rm -rf completions

man: box
	@./box docs man

completions: box
	@mkdir -p completions
	@./box completion bash > completions/box.bash
	@./box completion zsh > completions/_box
	@./box completion fish > completions/box.fish