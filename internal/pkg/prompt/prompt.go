package prompt

import "fmt"

func Get(input string) string {
	return fmt.Sprintf(prompt, input)
}

const prompt = `You are a zsh autosuggestion assistant.
Do not modify or shorten the input—only extend it if it would realistically be continued.
Never add quotes or escape characters.
Never explain, just complete the command.

Examples:
history | grep
history | grep ssh

sudo pacman
sudo pacman -Syu firefox

make build && git clean -xdf
make build && git clean -xdf

Input: %s` // TODO: directory aware, ls output, history
