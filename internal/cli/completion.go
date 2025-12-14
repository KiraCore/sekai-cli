// Package cli provides shell completion generators for sekai-cli.
package cli

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateBashCompletion generates a bash completion script for the given root command.
func GenerateBashCompletion(root *Command) string {
	var sb strings.Builder

	sb.WriteString(`#!/bin/bash
# sekai-cli bash completion script
# Generated automatically - do not edit

_sekai_cli_completions() {
    local cur prev words cword
    if type _init_completion &>/dev/null; then
        _init_completion || return
    else
        COMPREPLY=()
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
        words=("${COMP_WORDS[@]}")
        cword=$COMP_CWORD
    fi

    # Build the command path from words
    local cmd_path=""
    local i
    for ((i=1; i<cword; i++)); do
        if [[ "${words[i]}" != -* ]]; then
            if [[ -n "$cmd_path" ]]; then
                cmd_path="${cmd_path}_${words[i]}"
            else
                cmd_path="${words[i]}"
            fi
        fi
    done

    # Handle flag completion
    if [[ "${cur}" == -* ]]; then
        local flags=""
        case "$cmd_path" in
`)

	// Generate case statements for each command path
	walkCommandsForBash(root, "", &sb)

	sb.WriteString(`            *)
                flags="--help --config --output --container --node --chain-id --keyring-backend --home --rest"
                ;;
        esac
        COMPREPLY=($(compgen -W "${flags}" -- "${cur}"))
        return
    fi

    # Handle subcommand completion
    local commands=""
    case "$cmd_path" in
`)

	// Generate subcommand cases
	walkCommandsForBashSubcommands(root, "", &sb)

	sb.WriteString(`        *)
            commands=""
            ;;
    esac

    if [[ -n "$commands" ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
    fi
}

complete -F _sekai_cli_completions sekai-cli
`)

	return sb.String()
}

// walkCommandsForBash generates bash case statements for flag completion.
func walkCommandsForBash(cmd *Command, path string, sb *strings.Builder) {
	// Collect all flags for this command (including inherited)
	flags := collectAllFlags(cmd)
	if len(flags) > 0 {
		casePath := path
		if casePath == "" {
			casePath = `""`
		} else {
			casePath = fmt.Sprintf(`"%s"`, casePath)
		}

		sb.WriteString(fmt.Sprintf("            %s)\n", casePath))
		sb.WriteString(fmt.Sprintf("                flags=\"%s\"\n", strings.Join(flags, " ")))
		sb.WriteString("                ;;\n")
	}

	// Recurse into subcommands
	for _, sub := range cmd.SubCommands {
		if sub.Hidden {
			continue
		}
		subPath := sub.Name
		if path != "" {
			subPath = path + "_" + sub.Name
		}
		walkCommandsForBash(sub, subPath, sb)
	}
}

// walkCommandsForBashSubcommands generates bash case statements for subcommand completion.
func walkCommandsForBashSubcommands(cmd *Command, path string, sb *strings.Builder) {
	// Collect visible subcommands
	var subNames []string
	for _, sub := range cmd.SubCommands {
		if !sub.Hidden {
			subNames = append(subNames, sub.Name)
			// Add aliases too
			subNames = append(subNames, sub.Aliases...)
		}
	}

	if len(subNames) > 0 {
		sort.Strings(subNames)
		casePath := path
		if casePath == "" {
			casePath = `""`
		} else {
			casePath = fmt.Sprintf(`"%s"`, casePath)
		}

		sb.WriteString(fmt.Sprintf("        %s)\n", casePath))
		sb.WriteString(fmt.Sprintf("            commands=\"%s\"\n", strings.Join(subNames, " ")))
		sb.WriteString("            ;;\n")
	}

	// Recurse into subcommands
	for _, sub := range cmd.SubCommands {
		if sub.Hidden {
			continue
		}
		subPath := sub.Name
		if path != "" {
			subPath = path + "_" + sub.Name
		}
		walkCommandsForBashSubcommands(sub, subPath, sb)
	}
}

// collectAllFlags returns all flag names for a command (long and short forms).
func collectAllFlags(cmd *Command) []string {
	var flags []string
	seen := make(map[string]bool)

	// Add help flag always
	flags = append(flags, "--help", "-h")
	seen["help"] = true

	// Collect flags from current command
	for _, f := range cmd.Flags {
		if f.Hidden {
			continue
		}
		if !seen[f.Name] {
			flags = append(flags, "--"+f.Name)
			seen[f.Name] = true
		}
		if f.Short != "" && !seen[f.Short] {
			flags = append(flags, "-"+f.Short)
			seen[f.Short] = true
		}
	}

	return flags
}

// GenerateZshCompletion generates a zsh completion script for the given root command.
func GenerateZshCompletion(root *Command) string {
	var sb strings.Builder

	sb.WriteString(`#compdef sekai-cli
# sekai-cli zsh completion script
# Generated automatically - do not edit

_sekai-cli() {
    local curcontext="$curcontext" state line
    typeset -A opt_args

    local -a commands
    local -a global_flags

    global_flags=(
        '(-h --help)'{-h,--help}'[Show help]'
        '(-c --config)'{-c,--config}'[Path to config file]:file:_files'
        '(-o --output)'{-o,--output}'[Output format (text, json, yaml)]:format:(text json yaml)'
        '--container[Docker container name]:container:'
        '--node[Node RPC endpoint]:endpoint:'
        '--chain-id[Chain ID]:chain:'
        '--keyring-backend[Keyring backend]:backend:(test file os)'
        '--home[Sekaid home directory]:directory:_directories'
        '--rest[REST API endpoint]:endpoint:'
    )

    _arguments -C \
        $global_flags \
        '1: :->command' \
        '*:: :->args'

    case $state in
        command)
            commands=(
`)

	// Add root subcommands
	for _, sub := range root.SubCommands {
		if sub.Hidden {
			continue
		}
		desc := strings.ReplaceAll(sub.Short, "'", "'\\''")
		sb.WriteString(fmt.Sprintf("                '%s:%s'\n", sub.Name, desc))
	}

	sb.WriteString(`            )
            _describe -t commands 'sekai-cli commands' commands
            ;;
        args)
            case $line[1] in
`)

	// Generate cases for each subcommand
	for _, sub := range root.SubCommands {
		if sub.Hidden {
			continue
		}
		generateZshSubcommand(sub, &sb, 4)
	}

	sb.WriteString(`            esac
            ;;
    esac
}

_sekai-cli "$@"
`)

	return sb.String()
}

// generateZshSubcommand recursively generates zsh completion for a subcommand.
func generateZshSubcommand(cmd *Command, sb *strings.Builder, indent int) {
	prefix := strings.Repeat(" ", indent*4)

	sb.WriteString(fmt.Sprintf("%s%s)\n", prefix, cmd.Name))

	if len(cmd.SubCommands) > 0 {
		sb.WriteString(fmt.Sprintf("%s    local -a subcmds\n", prefix))
		sb.WriteString(fmt.Sprintf("%s    subcmds=(\n", prefix))
		for _, sub := range cmd.SubCommands {
			if sub.Hidden {
				continue
			}
			desc := strings.ReplaceAll(sub.Short, "'", "'\\''")
			sb.WriteString(fmt.Sprintf("%s        '%s:%s'\n", prefix, sub.Name, desc))
		}
		sb.WriteString(fmt.Sprintf("%s    )\n", prefix))
		sb.WriteString(fmt.Sprintf("%s    _describe -t commands '%s subcommands' subcmds\n", prefix, cmd.Name))
	} else {
		sb.WriteString(fmt.Sprintf("%s    _arguments $global_flags\n", prefix))
	}

	sb.WriteString(fmt.Sprintf("%s    ;;\n", prefix))
}

// GenerateFishCompletion generates a fish completion script for the given root command.
func GenerateFishCompletion(root *Command) string {
	var sb strings.Builder

	sb.WriteString(`# sekai-cli fish completion script
# Generated automatically - do not edit

# Disable file completion by default
complete -c sekai-cli -f

# Global flags
complete -c sekai-cli -s h -l help -d 'Show help'
complete -c sekai-cli -s c -l config -d 'Path to config file' -r
complete -c sekai-cli -s o -l output -d 'Output format' -xa 'text json yaml'
complete -c sekai-cli -l container -d 'Docker container name' -r
complete -c sekai-cli -l node -d 'Node RPC endpoint' -r
complete -c sekai-cli -l chain-id -d 'Chain ID' -r
complete -c sekai-cli -l keyring-backend -d 'Keyring backend' -xa 'test file os'
complete -c sekai-cli -l home -d 'Sekaid home directory' -ra '(__fish_complete_directories)'
complete -c sekai-cli -l rest -d 'REST API endpoint' -r

`)

	// Generate completions for all commands recursively
	walkCommandsForFish(root, "", &sb)

	return sb.String()
}

// walkCommandsForFish recursively generates fish completions.
func walkCommandsForFish(cmd *Command, path string, sb *strings.Builder) {
	// Generate subcommand completions
	for _, sub := range cmd.SubCommands {
		if sub.Hidden {
			continue
		}

		desc := strings.ReplaceAll(sub.Short, "'", "\\'")

		if path == "" {
			// Root level commands
			sb.WriteString(fmt.Sprintf("complete -c sekai-cli -n '__fish_use_subcommand' -a '%s' -d '%s'\n",
				sub.Name, desc))
		} else {
			// Nested commands - need to check parent command was seen
			sb.WriteString(fmt.Sprintf("complete -c sekai-cli -n '__fish_seen_subcommand_from %s; and not __fish_seen_subcommand_from %s' -a '%s' -d '%s'\n",
				path, sub.Name, sub.Name, desc))
		}

		// Add aliases
		for _, alias := range sub.Aliases {
			if path == "" {
				sb.WriteString(fmt.Sprintf("complete -c sekai-cli -n '__fish_use_subcommand' -a '%s' -d '%s (alias)'\n",
					alias, desc))
			}
		}

		// Recurse
		subPath := sub.Name
		if path != "" {
			subPath = path + " " + sub.Name
		}
		walkCommandsForFish(sub, subPath, sb)
	}

	// Generate flag completions for this command level
	if path != "" {
		for _, f := range cmd.Flags {
			if f.Hidden {
				continue
			}
			desc := strings.ReplaceAll(f.Usage, "'", "\\'")

			var flagDef string
			if f.Short != "" {
				flagDef = fmt.Sprintf("-s %s -l %s", f.Short, f.Name)
			} else {
				flagDef = fmt.Sprintf("-l %s", f.Name)
			}

			sb.WriteString(fmt.Sprintf("complete -c sekai-cli -n '__fish_seen_subcommand_from %s' %s -d '%s'\n",
				path, flagDef, desc))
		}
	}
}
