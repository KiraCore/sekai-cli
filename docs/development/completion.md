# Shell Completion

## Overview

Add shell completion support to sekai-cli for bash, zsh, and fish shells. This enables tab-completion for commands, subcommands, and flags.

## Implementation Tasks

| Task | Status |
|------|--------|
| Add `completion` command to CLI | |
| Implement bash completion generator | |
| Implement zsh completion generator | |
| Implement fish completion generator | |
| Add installation instructions to help | |
| Test bash completion | |
| Test zsh completion | |
| Test fish completion | |

## Command Structure

```
sekai-cli completion bash   # Output bash completion script
sekai-cli completion zsh    # Output zsh completion script
sekai-cli completion fish   # Output fish completion script
```

## Implementation Details

### Location
- `internal/app/app.go` - Add `buildCompletionCommand()` method
- `internal/cli/completion.go` - Completion script generators

### Bash Completion Generator

The generator should:
1. Walk the command tree recursively via `*cli.Command`
2. Generate `complete -F` statements for each command path
3. Handle flag completion (both `--long` and `-s` short forms)
4. Handle subcommand completion
5. Support dynamic completion hints for arguments (optional)

### Key Data Available from cli.Command

```go
type Command struct {
    Name        string      // Command name
    Aliases     []string    // Alternative names
    SubCommands []*Command  // Nested commands
    Flags       []Flag      // Command flags
    Args        []Arg       // Positional arguments
}

type Flag struct {
    Name    string  // Long flag name (without --)
    Short   string  // Short flag (single char, without -)
    Usage   string  // Description
}
```

### Bash Script Template

```bash
_sekai_cli_completions() {
    local cur prev words cword
    _init_completion || return

    local commands="status keys bank query tx config init sync cache version completion"

    case "${prev}" in
        sekai-cli)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            return
            ;;
        keys)
            COMPREPLY=($(compgen -W "add delete export import list show rename mnemonic" -- "${cur}"))
            return
            ;;
        # ... more cases
    esac

    # Flag completion
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help --config --output --container --node --chain-id" -- "${cur}"))
        return
    fi
}

complete -F _sekai_cli_completions sekai-cli
```

## User Installation

### Bash
```bash
# Option 1: System-wide (requires sudo)
sekai-cli completion bash | sudo tee /etc/bash_completion.d/sekai-cli

# Option 2: User-specific
sekai-cli completion bash >> ~/.bashrc
source ~/.bashrc
```

### Zsh
```bash
sekai-cli completion zsh > "${fpath[1]}/_sekai-cli"
# or
echo 'source <(sekai-cli completion zsh)' >> ~/.zshrc
```

### Fish
```bash
sekai-cli completion fish > ~/.config/fish/completions/sekai-cli.fish
```

## Testing

### Manual Testing
```bash
# Generate and source completion
source <(./build/sekai-cli completion bash)

# Test completion
sekai-cli <TAB><TAB>           # Should show: status keys bank query tx ...
sekai-cli keys <TAB><TAB>      # Should show: add delete export import list ...
sekai-cli --<TAB><TAB>         # Should show: --help --config --output ...
sekai-cli -<TAB><TAB>          # Should show: -h -c -o ...
```

### Integration Test
```go
func TestCompletionBash(t *testing.T) {
    // Run completion command
    // Verify output contains expected bash functions
    // Verify all root commands are listed
}
```

## References

- Bash completion: https://www.gnu.org/software/bash/manual/html_node/Programmable-Completion.html
- Zsh completion: https://zsh.sourceforge.io/Doc/Release/Completion-System.html
- Fish completion: https://fishshell.com/docs/current/completions.html
