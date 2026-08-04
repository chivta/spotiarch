# Personal Preferences

## Commands
- Don't run dev server commands (e.g., `air run`) -- assume it's already running in dev docker container with hot reload. If not start it with docker compose up -d.
- Don't run build commands unless specifically told to.

## Restrictions
- Do not ever try to ssh into remote machines.
- Do not push the code.
- Do not try to deploy.

## Code Style
- Always strive for concise, simple solutions.
- If a problem can be solved in a simpler way, propose it.
- Use coding style and architectural decisions similar to project ruscan accessible at @../ruscan. Use same deployment style with k8s directory, github actions etc. Use same commit style with concise one line explanations. NEVER add Co-authored line to commit message.

## General preferences
- If asked to do too much work at once, stop and state that clearly.

Model selection and Codex CLI usage live in the `codex-use` skill, not in this file.