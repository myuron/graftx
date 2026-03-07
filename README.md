# graftx

A TUI file manager that accelerates file copying between repositories.

[日本語版 README](README_ja.md)

With a dual-pane layout displaying Source and Destination side by side, you can quickly perform file operations using only the keyboard with Vim-like keybindings.

https://github.com/user-attachments/assets/4cb2b8a1-3310-4463-a375-02996ce8f257

## Requirements

- Go 1.21 or later
- [ghq](https://github.com/x-motemen/ghq) — Used to list repositories

## Installation

```bash
go install github.com/myuron/graftx@latest
```

## Getting Started

```bash
# Launch with the current directory as the right pane (Destination)
graftx
```

On launch, the right pane displays the contents of the current directory. Press `s` to select a source repository, which will be shown in the left pane.

## Fonts

File tree icons are supported via [Nerd Fonts](https://www.nerdfonts.com/). When using a terminal with Nerd Fonts installed, file-type-specific icons are displayed.

> The application works without Nerd Fonts, but icons may not render correctly.

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` | Move cursor down |
| `k` | Move cursor up |
| `l` | Enter directory |
| `h` | Go to parent directory |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `Tab` | Switch focus between left and right panes |

### File Selection

| Key | Action |
|-----|--------|
| `Space` | Toggle selection on cursor line (cursor moves down after selection) |
| `Ctrl+a` | Select all entries |
| `Ctrl+r` | Invert selection |
| `Esc` / `Ctrl+c` | Clear selection / Clear filter |

### Copy (Yank & Paste)

| Key | Action |
|-----|--------|
| `y` | Yank selected entries (or cursor line if none selected). Press `y` again in the same pane to cancel yank |
| `p` | Paste yanked entries to the focused pane (skip existing files) |
| `P` | Paste yanked entries with overwrite |

### File Operations

| Key | Action |
|-----|--------|
| `a` | Create new file/directory (append `/` for directory) |
| `r` | Rename |
| `d` | Move to trash (confirmation: `y`/`n`) |
| `D` | Delete permanently (confirmation: `y`/`n`) |

### Search & Filter

| Key | Action |
|-----|--------|
| `/` | Forward search mode (press Enter to confirm, matches are highlighted) |
| `n` | Jump to next search result |
| `N` | Jump to previous search result |
| `f` | Filter mode (show only entries containing the input string) |
| `Esc` / `Ctrl+c` | Clear filter |

### Repository Selection

| Key | Action |
|-----|--------|
| `s` | Open repository selection popup (lists repositories managed by ghq) |

Keys available in the popup:

| Key | Action |
|-----|--------|
| Type characters | Filter (incremental search) |
| `Ctrl+j` / `Ctrl+n` | Move cursor down |
| `Ctrl+k` / `Ctrl+p` | Move cursor up |
| `Enter` | Confirm selection |
| `Esc` / `Ctrl+c` | Cancel |

### Other

| Key | Action |
|-----|--------|
| `.` | Toggle hidden files visibility |
| `q` | Quit |

## Basic Usage

1. Run `graftx` (in the directory you want as the copy destination)
2. Press `s` to select the source repository
3. Navigate to the files you want to copy using `h`/`j`/`k`/`l` in the left pane
4. Select with `Space` (multiple selections allowed), then `y` to yank
5. Press `Tab` to switch to the right pane and navigate to the destination directory
6. Press `p` to paste

## License

MIT
