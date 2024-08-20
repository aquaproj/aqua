_aqua_env() {
    eval "$(aqua output-shell)"
}
autoload -Uz add-zsh-hook
add-zsh-hook preexec _aqua_env
