# lcd -- lupan's change directory

`lcd` is a tool to easily find a directory in your tree of deeply
nested source projects (i.e., with less typing).


## Installation

To install lcd you have to run

```
$ go get github.com/lukpank/lcd/cmd/lcd
```


## Configuration

1. First (if you are using `bash`) you have to add the following to
   your `~/.bashrc` file

   ```
   lcd() {
       local paths
       paths=$("$HOME/go/bin/lcd" "$@")
       if [ $? -eq 0 ]; then
           cd "${paths}"
       elif [ ! -z "${paths}" ]; then
           echo "${paths}"
       fi
   }

   _lcd() {
       local cur="${COMP_WORDS[COMP_CWORD]}"
       COMPREPLY=( $("$HOME/go/bin/lcd" -complete "${cur}") )
   }

   complete -F _lcd lcd
   ```

   And restart your running terminals or paste and execute the above
   declarations your running terminals.

2. Next, create the directory `~/.lcd` and create executable script
   `~/bin/lcd-update` (assumes `~/bin` is in your `PATH`) with the
   following content (that is just an example)

   ```
   #!/bin/sh

   exec find ~/src ~/go/src -type d | grep -v '/[.]git/' > ~/.lcd/cache
   ```

   And run the script with

   ```
   $ lcd-update
   ```

   You will have to rerun this script from time to time when you
   create new directories in `~/src` and `~/go/src` (or whatever your
   selected directories will be).


## Usage

Project paths can be long and inconvenient to type.

1. For example to enter to `guru` command directory with `cd`
   (provided your GOPATH is `~/go`) you have to write

   ```
   ~$ cd ~/go/src/golang.org/x/tools/cmd/guru
   guru$ █
   ```

   with `lcd` you just write

   ```
   $ lcd guru
   /home/user/go/src/golang.org/x/tools/cmd/guru
   guru$ █
   ```

   that was a simplest case: as there were only single directory
   having the last component `guru` (everythig after last `/`) thus
   `lcd` enters to this particular directory.

2. But if we want to enter into the `lcd` directory we write

   ```
   ~$ lcd lcd
   Search: █
     Change directory:
     ▸ /home/user/go/src/github.com/lukpank/lcd
       /home/user/go/src/github.com/lukpank/lcd/cmd/lcd
   ```

   and `lcd` displays a menu of possible directories (here only two).
   You can navigate to the intended directory with arrows (`↓` and
   `↑`) or Emacs next and previous line keys (`Ctrl+n` and `Ctrl+p`)
   and select the intended directory with `Enter`.  You can also use
   `←` and `→` or `Ctrl+b` and `Ctrl+f` to jump to previous and next
   page, respectively.  `lcd` starts in search mode (described below),
   you can leave the search mode with `Ctrl+U` and then you can
   additionally use vim keys `j`, `k`, `h` and `l` to move in the
   menu.


3. If we want to enter the second of the above directories we can
   write

   ```
   $ lcd cmd/lcd
   /home/user/go/src/github.com/lukpank/lcd/cmd/lcd
   lcd$ █
   ```

   and avoid the selection with menu (as there is only one path ending
   with `cmd/lcd`).

4. But if we recently searched for the `lcd` directory we may also
   remember that the intended directory was the second item of the
   menu so we can write

   ```
   $ lcd lcd 2
   /home/user/go/src/github.com/lukpank/lcd/cmd/lcd
   lcd$ █
   ```

   to enter to the second of the directories. (note: the order may
   change if we updated the cache with `lcd-update` or deleted some of
   the directories).

5. We configured `lcd` with bash completions so we can write

   ```
   ~$ lcd glpk<TAB><TAB>
   glpk       glpk-4.65
   ~$ lcd glpk█
   ```

   and after hitting `TAB` twice the list of possible completion of
   the `glpk` will be displayed (here we do have two directories
   having last component starting with `glpk`).

6. We could also complete paths for which one before last element is
   `cmd` by writing

   ```
   $ lcd cmd/<TAB><TAB>█
   Display all 101 possibilities? (y or n)
   ```

   there are 101 such paths so we can list them all or try to be more
   specific by providing at least one letter of the intended final
   directory component.

7. As mentioned before, `lcd` menu starts in search mode so after
   giving the final component of the path on the command line you can
   also filter with one of more words like below:

   ```
   $ lcd cmd
   Search: github.com golang█
     Change directory:
     ▸ /home/user/go/src/github.com/golang/freetype/cmd
       /home/user/go/src/github.com/golang/dep/cmd
       /home/user/go/src/github.com/golang/dep/gps/_testdata/cmd
   ```

   Only paths containing all of the words as substrings will be
   displayed in the menu. You can prepend a word with `!` to filter
   out given word from the menu.  Hitting `Ctrl+U` twice leaves and
   reenters the search mode and clears the search string.


## License

MIT


## See also

There is a general purpose console menu with fuzzy search called
[skim](https://crates.io/crates/skim) (written in Rust).
