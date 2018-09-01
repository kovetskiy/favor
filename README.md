# favor

not production ready, but you can test it already

```
bindkey -v '^N' :favor
zle -N :favor
:favor() {
    local favor_dir="$(favor 2>/dev/null)"
    if [[ ! "$favor_dir" ]]; then
        return
    fi

    eval cd "$favor_dir"
    unset favor_dir

    clear
    zle -R
    # uncomment for lambda17 prompt compatibility
    # lambda17:update
    zle reset-prompt
}
```
