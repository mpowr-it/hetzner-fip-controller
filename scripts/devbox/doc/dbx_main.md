# DevBox Shell | MPOWR-OSS-ENV

_This shell contains all tools and scripts for the planned production use of the DEVBOX as an integral part of our project repository structure. Extensions and customisations are not only desired but also allowed at any time - we therefore ask for corresponding merge requests for the package and script section of the `devbox.json` file as well as this documentation._

## Available Scripts

```bash
# call devbox-linked script|action
$ devbox run <script>
# e.g. devbox run help
```

| Script       | Description                                                                            |
|--------------|:---------------------------------------------------------------------------------------|
| `init-pyenv` | initialize a dedicated python3 environment (experimental feature)                      |
| `help`       | print-out this markdown file using glow + bat to show available scripts and alias-sets |

## Terminal Usage

During your active DevBox-Terminal session, your previous alias commands and shell settings are fully preserved, which means that you should still have access to your terminal-relevant configurations and application facilitation's.

## TMUX Help

If you are in an active tmux session within the devbox, you can display a short help on TMUX commands using the following command line

```bash
PAGER='bat' glow -p scripts/devbox/doc/dbx_tmux.md
```

## Links

[![WIKI](https://img.shields.io/badge/Confluence%2FDocs-black)](#)
[![MIRO](https://img.shields.io/badge/MIRO-DevBox%2FIntro-lightgray)](#)
[![REPO](https://img.shields.io/badge/Core%20Version-1.0.0-green.svg)](#)
