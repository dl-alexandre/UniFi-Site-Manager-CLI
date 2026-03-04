# Bash completion for usm (UniFi Site Manager CLI)
# Source this file in your ~/.bashrc or ~/.bash_profile

_usm_complete() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Main commands
    local commands="account site device network config help --version --help"
    
    # Account subcommands
    local account_cmds="login logout status"
    
    # Site subcommands
    local site_cmds="list create delete select current"
    
    # Device subcommands
    local device_cmds="list adopt restart upgrade"
    
    # Network subcommands
    local network_cmds="status clients"
    
    # Config subcommands
    local config_cmds="init view set"
    
    case "${prev}" in
        account)
            COMPREPLY=( $(compgen -W "${account_cmds}" -- ${cur}) )
            return 0
            ;;
        site)
            COMPREPLY=( $(compgen -W "${site_cmds}" -- ${cur}) )
            return 0
            ;;
        device)
            COMPREPLY=( $(compgen -W "${device_cmds}" -- ${cur}) )
            return 0
            ;;
        network)
            COMPREPLY=( $(compgen -W "${network_cmds}" -- ${cur}) )
            return 0
            ;;
        config)
            COMPREPLY=( $(compgen -W "${config_cmds}" -- ${cur}) )
            return 0
            ;;
        *)
            ;;
    esac
    
    # Complete main commands
    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "--version --help" -- ${cur}) )
    else
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
    fi
}

complete -F _usm_complete usm
