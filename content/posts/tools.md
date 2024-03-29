---
title: "Developer Tools"
date: 2022-03-09T09:49:24Z
draft: false
categories:
  - "tools"
  - "guide"
tags:
  - "shell"
  - "terminal"
  - "vim"
---

I don't know who needs to hear this but:
**I love my developer tools**.

## Core

- **Package Manager**: [Homebrew](https://brew.sh/)
- **Terminal**: [Warp](https://www.warp.dev/)
- **Code Editor**: [Neovim](https://neovim.io)

**NOTES:**  
I was using a multiplexer ([tmux](https://github.com/tmux/tmux/wiki)) for years but colour issues forced me to switch to [Zellij](https://zellij.dev). I then switched from [Alacritty](https://alacritty.org/) and Zellij to using Warp as my terminal (helping to consolidate the number of tools I needed to install and configure) as Warp provided a lot of what I was using Zellij for. My switch to Warp also meant I could drop support for [FZF](https://github.com/junegunn/fzf/blob/master/README.md) as Warp provided fuzzy searching my history, while `fd` (see below) allowed me to easily locate files.

## Supplementary

- **Code Statistics**: [Tokei](https://github.com/XAMPPRocky/tokei/blob/master/README.md)
- **DNS Client**: [Dog](https://github.com/ogham/dog/blob/master/README.md)
- **Directory Lister**: [Tree](https://en.wikipedia.org/wiki/Tree_(command))
- **Directory Navigator**: [Broot](https://github.com/Canop/broot/blob/master/README.md)
- **Directory Switcher**: [Zoxide](https://github.com/ajeetdsouza/zoxide/blob/main/README.md)
- **Disk Usage**: [Duf](https://github.com/muesli/duf/blob/master/README.md)
- **Feed Reader**: [Tuifeed](https://github.com/veeso/tuifeed/blob/main/README.md)
- **File Content Display**: [Bat](https://github.com/sharkdp/bat/blob/master/README.md)
- **File Finder**: [Fd](https://github.com/sharkdp/fd/blob/master/README.md)
- **File Lister**: [Exa](https://github.com/ogham/exa/blob/master/README.md)
- **File Removal**: [Rip](https://github.com/nivekuil/rip/blob/master/README.org)
- **File Space Usage**: [Dust](https://github.com/bootandy/dust/blob/master/README.md)
- **Network Reachability**: [Gping](https://github.com/orf/gping/blob/master/readme.md)
- **Network Utilisation**: [Bandwhich](https://github.com/imsnif/bandwhich/blob/main/README.md)
- **Process Status**: [Procs](https://github.com/dalance/procs/blob/master/README.md)
- **Process/System Monitor**: [Htop](https://github.com/htop-dev/htop/#readme) †
- **Shell**: [Zsh](https://www.zsh.org)
- **Shell Autocomplete**: [Fig](https://fig.io/)
- **Shell Prompt**: [Starship](https://starship.rs/)

> † I had switched to [Bottom](https://github.com/ClementTsang/bottom/blob/master/README.md) but `htop` is just so much more flexible and configurable, and although the graphs in `btm` are nice I just don't _need_ them.

## Browser

Controversial, but I had started to move away from [Firefox](https://www.mozilla.org/en-GB/firefox/new/) with a whole suite of security/privacy minded tools to [Apple's Safari](https://www.apple.com/uk/safari/) browser.

I _was_ using Safari's ["Tab Group"](https://twitter.com/integralist/status/1514526555275501569?s=20&t=BJu3WlWq6dhoeAarJf91ig) feature, which I _really_ liked along with the following extensions...

- AdBlock One
- Dark Reader for Safari
- SimplyJSON for Safari
- Super Agent for Safari (Cookie Consent Automation)

...but ultimately the Tab Groups feature ended up being super buggy and duplicating bookmarks and groups etc and I couldn't get the issue resolved, so I ended up back with Google Chrome.
