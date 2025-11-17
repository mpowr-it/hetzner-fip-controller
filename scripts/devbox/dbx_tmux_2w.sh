#!/usr/bin/env bash
# --
# @version: 1.0.0
# @purpose: Shell script to run when an interactive tmux/devbox-based shell for ice api tunnel terminal support
# --

if tmux has-session -t $C_DBX_INIT_SESSION_NAME_2W 2>/dev/null; then
    echo "DevBox TMUX-Session [$C_DBX_INIT_SESSION_NAME_2W] already exists, type exit on all panes to return to 'normal' devbox shell ..."
    exit 0
else
    # init: set file-marker for active tmux session; this will be used to handle additional help output during init-phase of tmux panes
    touch $C_DBX_INIT_SESSION_FILE_MARKER >/dev/null; echo "DevBox TMUX-Session [$C_DBX_INIT_SESSION_NAME_2W] prepared"
    # use tmuxp to initialize/bootstrap the tmux session used for our double-pane terminal
    tmuxp load ./scripts/devbox/init/tmuxp_2w.yaml
    # de-init: remove file-marker for past session process; prevent help-output un devbox terminal normal session(s)
    rm -f $C_DBX_INIT_SESSION_FILE_MARKER >/dev/null; echo "DevBox TMUX-Session [$C_DBX_INIT_SESSION_NAME_2W] cleared"
fi
