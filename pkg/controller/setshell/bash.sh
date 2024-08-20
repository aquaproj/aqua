if [ -n "${PROMPT_COMMAND:-}" ]; then
    PROMPT_COMMAND="${PROMPT_COMMAND};"'eval "$(aqua output-shell)"'
else
    PROMPT_COMMAND='eval "$(aqua output-shell)"'
fi
