---
title: "New Laptop Configuration (Second Edition)"
date: 2022-05-26T13:33:32+01:00
categories:
  - "code"
  - "development"
tags:
  - "configuration"
  - "laptop"
  - "new"
draft: false
---

## Introduction

This is the second edition to ["New Laptop Configuration"](/posts/new-laptop-configuration/).

## Backup

```bash
cd ~/
mkdir /tmp/keys
gpg --export-secret-keys --armor <NAME> > /tmp/keys/<NAME>.asc
gpg --symmetric /tmp/keys/<NAME>.asc
gpg --export-ownertrust > /tmp/keys/trustdb.txt 

zip -r /tmp/keys/sshbackup ~/.ssh/
unzip -l /tmp/keys/sshbackup.zip
gpg --symmetric /tmp/keys/sshbackup.zip

rm -rf /tmp/keys
```

## Steps

1. Install Rust.
  ```bash
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
  ```
2. Install Go.
  ```txt
  https://go.dev/dl/
  ```
3. Import GPG/SSH keys.
  ```bash
  mkdir /tmp/keys
  cd /tmp/keys
  
  gpg --decrypt /tmp/keys/<NAME>.gpg > <NAME>
  gpg --import <NAME>.asc
  
  rm ~/.gnupg/trustdb.gpg
  gpg --import-ownertrust < /tmp/keys/trustdb.txt
  
  gpg --decrypt /tmp/keys/sshbackup.zip.gpg > sshbackup.zip
  unzip /tmp/keys/sshbackup.zip
  mv /tmp/keys/.ssh/ ~/
  
  rm -rf /tmp/keys
  ```
4. Setup SSH.
  ```bash
  eval "$(ssh-agent -s)"
  ssh-add --apple-use-keychain ~/.ssh/github
  ```
5. Install Alacritty.
  ```bash
  # https://github.com/alacritty/alacritty/releases
  mkdir .bash_completion
  curl https://raw.githubusercontent.com/alacritty/alacritty/master/extra/completions/alacritty.bash -o ~/.bash_completion/alacritty
  ```
6. Install Fig.
  ```txt
  https://fig.io/
  ```
7. Install Homebrew + packages.
  ```bash
  # https://brew.sh
  brew bundle --file /tmp/Brewfile install
  ```
8. Change default shell to Zsh.
  ```bash
  echo /opt/homebrew/bin/zsh | sudo tee -a /etc/shells
  chsh -s /opt/homebrew/bin/zsh
  ```
9. Configure Vim-Plug.
  ```bash
  sh -c 'curl -fLo "${XDG_DATA_HOME:-$HOME/.local/share}"/nvim/site/autoload/plug.vim --create-dirs \
       https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim'
  ```
10. Setup dotfiles.
  > **NOTE**: Don't forget to execute 'prefix + I' in tmux to install plugins.
  ```bash
  cd /tmp
  git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
  git clone https://github.com/Integralist/dotfiles.git
  curl https://raw.githubusercontent.com/git/git/master/contrib/completion/git-prompt.sh -o ~/.git-prompt.sh
  curl https://raw.githubusercontent.com/git/git/master/contrib/completion/git-completion.zsh -o ~/.zsh/_git
  cp -R .alacritty.yml .zsh .zshrc .config .gitconfig .gitignore .gnupg .ignore .inputrc .leptonrc .tmux.conf ~/
  chown -R $(whoami) ~/.gnupg/
  chmod 600 ~/.gnupg/*
  chmod 700 ~/.gnupg
  ```
11. Setup password store.
  ```bash
  KEY_ID=$(gpg --list-keys <NAME> | head -n 2 | tail -n 1 | cut -d ' ' -f 7)
  pass init $KEY_ID
  pass git init
  pass git remote add origin git@github.com:<private/repo>
  pass git pull
  ```
12. Safari extensions.
  ```txt
  AdBlock One
  Dark Reader for Safari
  Super Agent for Safari (Cookie Consent Automation)
  Tab Sessions
  ```
13. Configure OS.
  ```txt
  - Dock (Automatically hide and show the Dock)
  - Keyboard (Key Repeat = Fast, Delay Until Repeat = Short)
  - Accessibility > Zoom (Use keyboard shortcuts to zoom)
  - Date & Time > Clock (Show date + Display the time with seconds)
  - Mission Control (disable "Automatically rearrange Spaces based on most recent use")
  - Terminal Developer Mode (`spctl developer-mode enable-terminal`)
  - Wake up from sleep (`sudo pmset -a standbydelay 7200`)
  ```
