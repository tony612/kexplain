# Kexplain

Kexplain is an interactive `kubectl explain`. It supports viewing the resource documentation
like `less` and jumping around between documentation of fields.

[![asciicast](https://asciinema.org/a/492648.svg)](https://asciinema.org/a/492648)


## Keys

| Key |      Action     |
| --- | ----------------- |
| <kbd>j</kbd> / <kbd>Ctrl-n</kbd> / <kbd>↓</kbd> | Move one line down  |
| <kbd>k</kbd> / <kbd>Ctrl-p</kbd>/ <kbd>↑</kbd>   | Move one line up  |
| <kbd>Tab</kbd>       | Select next field |
| <kbd>Shift</kbd>+<kbd>Tab</kbd> | Select previous field |
| <kbd>Alt-]</kbd> / <kbd>Alt</kbd>+<kbd>→</kbd> / <kbd>Enter</kbd>  | Go to the documentation of the selected field |
| <kbd>Alt-[</kbd> / <kbd>Alt</kbd>+<kbd>←</kbd>    | Go back to the previous documentation |
| <kbd>Ctrl-f</kbd> | Move one page down  |
| <kbd>Ctrl-b</kbd> | Move one page up  |
| <kbd>g</kbd>      | Move to the head  |
| <kbd>G</kbd>      | Move to the bottom  |
| <kbd>/</kbd>, type `word`, <kbd>Enter</kbd>    | Search `word`  |
| <kbd>n</kbd>      | Repeat previous search  |
| <kbd>N</kbd>      | Repeat previous search in reverse direction.  |
| <kbd>q</kbd> / <kbd>Q</kbd>  | Quit  |
