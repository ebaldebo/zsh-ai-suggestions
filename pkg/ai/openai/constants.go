package openai

const openAIURL = "https://api.openai.com/v1/chat/completions"

// Environment variables
const (
	envAPIKey = "ZSH_AI_SUGGESTIONS_OPENAI_API_KEY"
	envModel  = "ZSH_AI_SUGGESTIONS_OPENAI_MODEL"

	defaultModel = "gpt-4o-mini"
)

const (
	roleSystem = "system"
	roleUser   = "user"
)

const prompt = `You're a zsh autosuggestion assistant. Given a partial command, complete it with a realistic example command.
Don't use placeholders like 'pattern' or 'file.txt'. Provide a likely completion based on common command usage.

Examples:
Input: "grep"
Output: "grep -ri 'gorilla' ./documents/"

Input: "cat monkey.txt | grep"
Output: "cat monkey.txt | grep gorilla"

Input: "git checkout"
Output: "git checkout feature/authentication"

Respond ONLY with the completed command.` // TODO: directory aware, ls output, history
