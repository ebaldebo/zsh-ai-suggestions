package openai

const openAIURL = "https://api.openai.com/v1/chat/completions"

const (
	envAPIKey = "ZSH_AI_SUGGESTIONS_OPENAI_API_KEY"
	envModel  = "ZSH_AI_SUGGESTIONS_MODEL"

	defaultModel = "gpt-4o-mini"
)

const (
	roleSystem = "system"
	roleUser   = "user"
)
