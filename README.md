# zsh-ai-suggestions

## How to Run in Docker

1. Create a `.env` file in the root directory of the project with the following content:
   ```bash
   touch .env
   echo "ZSH_AI_SUGGESTIONS_TYPE=openai" >> .env
   echo "ZSH_AI_SUGGESTIONS_OPENAIAPI_KEY=your_openai_api_key" >> .env
   ```

2. Run using docker compose:
   ```bash
   make docker
   ```

3. In the container source the plugin:
   ```bash
   source /home/test/zsh-ai-suggestions.zsh
   ```

4. Type something and press Ctrl+/
   ```bash
   history |
   ```
