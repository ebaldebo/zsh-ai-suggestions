package gemini

const (
	envAPIKey = "ZSH_AI_SUGGESTIONS_GEMINI_API_KEY"
	envModel  = "ZSH_AI_SUGGESTIONS_MODEL"

	defaultModel = "gemini-2.0-flash-lite"

	baseUrl = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
)
