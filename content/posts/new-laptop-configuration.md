---
title: "New Laptop Configuration"
date: 2019-04-10T21:24:15+01:00
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

I'm an engineer with a _new_ laptop, which requires setting up with various development tools and configuration. This post is my attempt to capture and document my process for getting a new dev environment set-up.

I used to try and automate a lot of this with bash scripts, but realised over time that things go out of date quite quickly (e.g. OS configurations can change substantially, as well as my preferred ways of working). 

I also find that if an error occurs with an automated script (unless you've coded things defensively enough) you can end up with your machine in a weirdly broken state.

Given a straight forward set of instructions, doing things _manually_ doesn't take long at all, and you can modify things at that point in time without just blindly installing various things you no longer need.

Here's a list of what we're going to be discussing...

- [Defaults](#defaults)
- [Package Manager](#package-manager)
- [Essential Packages](#essential-packages)
- [Essential Apps](#essential-apps)
- [Curl](#curl)
- [Git](#git)
- [Shell](#shell)
- [Terminal](#terminal)
- [GitHub](#github)
- [Password Store](#password-store)
- [Go](#go)
- [Python](#python)
- [Vim](#vim)
- [Tmux](#tmux)
- [Miscellaneous](#miscellaneous)
- [macOS](#macos)
- [Homebrew Output](#homebrew-output)

## Defaults

It's good to begin by surveying your current system and understanding what you have already installed. For me this looked something like:

- OS: macOS Mojave (`10.14.4`)
- Curl: `/usr/bin/curl` (`7.54.0`)
- Bash: `/bin/bash` (`3.2.57`)
- Python: `/usr/bin/python` (`2.7.10`)
- Ruby: `/usr/bin/ruby` (`2.3.7p456`)
- Git: `/usr/bin/git` (`2.20.1`)
- $PATH: `/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin`

What's worth me noting additionally here is that I primarily use two programming languages: Go and Python. The reason I mention this is because Python has an interesting history with the name of its binaries. 

The binary name `python` generally refers to Python version `2.x`. Where as Python `3.x` has traditionally been named `python3` to help distinguish the two. So looking above we can see `which python` reveals the location as `/usr/bin/python` and without checking the version (e.g. `python --version`) I was fairly certain it would be a `2.x` version (based on the naming history).

This has been the generally accepted rule for a while, _except!_ when dealing with tools that handle virtual environments. 

For example, [pipenv](https://pipenv.readthedocs.io/) is a tool that helps you to manage not only different Python versions but also the dependencies installed for different projects (referred to as virtual environments). A tool like pipenv will proxy a command such as `python` through a shim script (e.g. `/Users/integralist/.pyenv/shims/python`) and that shim script will then determine which Python interpreter/binary to execute. 

A shim script typically identifies the virtual environment you're working under and will then figure out the most appropriate Python interpreter to invoke. So within that virtual env if you call `python`, then you may not necessarily get the Python2 interpreter, as your virtual env might be configured such that the expectation is to proxy your invocation to a Python3 interpreter.

This is why, when setting up a new laptop, getting a good development environment setup is essential because it can get quite confusing untangling a mess of default Python's vs `brew install ...` versions of Python 3 and then subsequently using multiple environment tools like `pipenv` which confuse things further by hiding the actual versions behind the generically named `python` command.

The situation reminds me a lot of [XKCD's classic comic strip](https://xkcd.com/1987/)...

<a href="../../images/python_env.png">
    <img src="../../images/python_env.png">
</a>

## Package Manager

Let's begin our journey by first installing a 'package manager'. This software will enable us to search and install various pieces of software. The macOS provides its own GUI implementation referred to as the 'App Store', but it's heavily moderated by Apple and an app can only be found there if it abides by Apple's own set of rules and criteria for what they consider to be 'safe' software.

> Note: there are many apps that aren't available in the App Store because Apple can be a bit anti-competition (see [Spotify's "time to play fair" campaign](https://timetoplayfair.com)).

So we have to download our own package manager, and the defacto standard for macOS is a program called [Homebrew](https://brew.sh/) (which is a terminal based tool, so no GUI). In fact, I'm not actually sure what _alternatives_ to Homebrew exist (other than [MacPorts](https://www.macports.org), which if you want to understand the differences between it and Homebrew then [read this](https://saagarjha.com/blog/2019/04/26/thoughts-on-macos-package-managers/))? On Linux you have tools such as `yum` or `apt` but for macOS you either use the built-in App Store or find your own alternative (so in this case, we'll use Homebrew).

To install Homebrew, execute the following command in your terminal:

```
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
```

> Note: notice that installation command uses the default installation of Ruby which Homebrew presumes is available (and for the most part is a safe presumption to make as Ruby as been provided by macOS for the longest time).

If you need to update Homebrew you can execute a `brew update` command, but the installation will install the latest version any way, so that won't be necessary.

## Essential Packages

OK, so I start with what I refer to as a 'essential packages', and specifically these are packages that do not require any configuration on my part. Meaning, I can install them and consider the job done, where as with other packages I install I'll have to make some additional tweaks to (which we'll see as we move on past the 'essential' segments of this post).

To install a package via Homebrew, execute the following command in your terminal:

```
brew install <package_name>
```

Here is a list of the packages I'll install:

- `ag`: a `grep` like tool (aka. the_silver_searcher)
- `gnu-sed`: it's the gnu version of `sed` (`gsed`) used for filtering/transforming text
- `jq`: tool for parsing/inspecting json output
- `docker`: useful for running containerized programs
- `hugo`: static site generator (used to make this website)
- `node`: server-side js programming language (used to compile a static search feature for my website)
- `pwgen`: generates random passwords
- `reattach-to-user-namespace`: used by tmux for clipboard storage
- `shellcheck`: bash linter
- `sift`: another command line search tool
- `transmission`: torrent client † (alt. `npm install -g t-get`)
- `tree`: displays directory heirarchy structures as a tree
- `watch`: executes given command every N seconds

> † see [transmission user guide](https://cli-ck.io/transmission-cli-user-guide/)

Here's a handy one-liner:

```
brew install ag gnu-sed jq docker hugo node pwgen reattach-to-user-namespace shellcheck sift tree watch
```

## Essential Apps

Homebrew now allows you to also install GUI applications, not just command line tools, but to do this you'll need to configure Homebrew to use `Cask`:

```
brew tap homebrew/cask
```

Once that's done you can install an app via Homebrew using the command:

```
brew cask install <app_name>
```

Here is a list of the apps I'll install:

- `alfred`: like Apple's Spotlight search, but better
- `caffeine`: stops the Mac from going to sleep
- `dash`: offline documentation
- `docker`: this is the counter-part to the 'package' installed earlier † 
- `google-backup-and-sync`: syncs files between computer and Google Drive
- `google-chrome`: web browser
- `keybase`: encryption tool
- `lepton`: GitHub Gist UI
- `slack`: communication tool
- `spotify`: music streaming service
- `vlc`: video player with support of lots of codecs

> † if you installed the docker 'package', then you _need_ the docker 'app' as well for it to work. You can't have one without the other (this is because the app sets up the interface for macOS to interact with the underlying docker client/server implementation).

Here's a handy one-liner:

```bash
brew cask install alfred caffeine dash docker google-backup-and-sync google-chrome keybase lepton slack spotify vlc
```

The [Dash](https://kapeli.com/dash) app will ask you what documentation you would like to download so it's available offline. I use the following docsets (I used to have _lots_ more but realised I never really used them, so this is my 'essential' docs list):

- boto3
- Go
- HTTP Header Fields
- HTTP Status Codes
- NGINX
- Python2
- Python3
- Regular Expressions
- tmux
- Tornado
- vim

## Curl

I like to use a more modern version of `curl` (e.g. supports HTTP/2, and other features):

```bash
brew install curl
```

But in order to use this version of curl you'll need to modify your `$PATH`:

```bash
export PATH="/usr/local/opt/curl/bin:$PATH"
```

## Git

Similarly to curl, I like to have the most recent version of `git` installed:

```bash
brew install git
```

Once that's installed I configure it like so:

```bash
curl -LSso ~/.git-prompt.sh https://raw.githubusercontent.com/git/git/master/contrib/completion/git-prompt.sh
curl -LSso ~/.gitignore-global https://raw.githubusercontent.com/Integralist/dotfiles/master/.gitignore-global
curl -LSso ~/.gitconfig https://raw.githubusercontent.com/Integralist/dotfiles/master/.gitconfig
```

> Note: it's always worth checking `~/.gitignore-global` is up to date (i.e. not referencing file types I no longer work with).

## Shell

To install and configure latest version of the Bash shell, execute the following commands:

```bash
brew install bash
echo /usr/local/bin/bash | sudo tee -a /etc/shells
chsh -s /usr/local/bin/bash
```

Also make sure to install auto-complete for bash:

```bash
brew install bash-completion
```

Finally, we'll configure Bash like so:

```bash
curl -LSso ~/.bash-preexec.sh https://raw.githubusercontent.com/rcaloras/bash-preexec/master/bash-preexec.sh
curl -LSso ~/.bashrc https://raw.githubusercontent.com/Integralist/dotfiles/master/.bashrc
curl -LSso ~/.bash_profile https://raw.githubusercontent.com/Integralist/dotfiles/master/.bash_profile
```

> Note: `~/.bashrc` references `~/.fzf.bash` which is needed, and comes from installing the FZF vim plugin (which we'll sort out shortly).

## Terminal

To install my custom terminal theme:

```bash
curl -LSso /tmp/Integralist.terminal https://raw.githubusercontent.com/Integralist/mac-os-terminal-theme-integralist/master/Integralist.terminal
open /tmp/Integralist.terminal
rm /tmp/Integralist.terminal
```

> Note: don't forget to change the terminal font to menlo (if not already set) and also set `Integralist` theme as your default. I used to do this via `defaults write com.apple.Terminal "Default Window Settings" Integralist` and `defaults write com.apple.Terminal "Startup Window Settings" Integralist` but those have changed now in the latest macOS (see `defaults read`).

## GitHub

Let's now set-up a new SSH key for GitHub access:

```bash
mkdir ~/.ssh
cd ~/.ssh && ssh-keygen -t rsa -b 4096 -C 'foobar@example.com'
eval "$(ssh-agent -s)"
ssh-add -K ~/.ssh/github_rsa
```

Don't forget to `pbcopy < ~/.ssh/github_rsa.pub` and paste your public key into the GitHub UI. Once that's done you can execute the following command to test your SSH key is set-up correctly and working:

```bash
ssh -T git@github.com
```

> Note: there is a slight catch-22 here which is if your password for GitHub is in your Password Store (see next section), then that makes things trickier. For me I also have a copy of the encrypted store on my phone and so I can utilise that to access the password. But failing that, you can just 'reset your password' in GitHub UI's and follow the email instructions to gain access and thus login and add your new SSH key.

## Password Store

I use the open-source [password store](https://www.passwordstore.org) for handling secrets and passwords. This tool provides the `pass` command, and that requires the use of `gpg`, so let's start by installing GPG:

```bash
brew install gpg
```

Now you have `gpg`, make sure you pull your private key from wherever you have it stored (e.g. external USB stick), then execute:

```bash
gpg --import <private-key>
gpg --export <key-id> # public key by default
```

> Note: don't forget you can _sign_ your git commits:  
> `git config --global user.signingkey <key-id>`  
> But you might prefer to use a Keybase key for that instead.

Next, install `pass` and `pass otp` commands:

```
brew install pass pass-otp zbar
```

You can now pass a QR code into `pass otp` and use the terminal for generating one-time pass codes for 2FA/MFA authentication:

```bash
zbarimg -q --raw /tmp/qr.png | pass otp insert Work/Acme/totp/foo`  

pass otp -c Work/Acme/totp/foo
```

> Note: installing `zbar` provides the `zbarimg` command

Lastly, we need to setup a new Password Store, and to do that we need to provide our GPG key id:

```bash
# <ref> needs to be your email, or part of the key's description
keyid=$(gpg --list-keys <ref> | head -n 2 | tail -n 1 | cut -d ' ' -f 7)

pass init $keyid
```

Now we can pull our Password Store from a _private_ repository:

```bash
pass git init
pass git remote add origin git@github.com:Foo/Bar.git
pass git pull
git branch --set-upstream-to=origin/master master
```

> Note: I also like to ensure my encrypted datastore is sync'ed up to other online providers, and symlinked to `~/.password-store` as well so changes are backed up automatically in multiple places.

### Mobile Password Store

There is also a [mobile app for Android](https://github.com/android-password-store/Android-Password-Store) that you can download from the Google Play store (and other places) that allows you to access the Password Store if it has been pushed to a distributed version-control system such as GitHub (better still if the repository is private -- "out of sight, out of mind").

To get set-up, go through the following steps:

- Install Password Store app.
- Select "Clone from Server" option.
- Add in github credentials (e.g. `git@github.com:Foo/Bar.git`)
- Create new SSH key via Password Store app (give it a password).
- Encrypt your SSH key with symmetrical encryption (e.g. `gpg --symmetric`)
- Email SSH key to self.
- Decrypt SSH key and copy it into GitHub UI.
- Password Store app will ask for SSH key password, then it'll clone the repo.

Before you can access the content of the Password Store (remember all the content is individual GPG keys) you'll need the GPG _private_ key in order to decrypt files that would have been encrypted using your public key.

- Export your private key as ASCII (via laptop: `gpg --armor --output passkey.txt --export-secret-keys <key_id>`).
- Encrypt your exported private key with symmetrical encryption (`gpg --symmetric passkey.txt`)
- Email your private key to yourself.
- Download private key to your phone.
- Install OpenKeychain app (via Google Play)
- Locate the downloaded encrypted private key and open with OpenKeychain app.
- Enter password used to encrypt the private key.
- Then click into the extracted `passkey.txt` file and then click "Import".
- In Password Store app set the options:
  - Select "OpenPGP Provider" (choose: OpenKeychain)
  - Select "OpenPGP Key id" (choose: the imported public key)

## Go

We'll install the latest version of `go` (as far as Homebrew is concerned):

```bash
brew install go
```

This is required because in order to handle different versions of `go` we'll want to [manually compile go](https://gist.github.com/af300f602fa4da8cc14863f36a24bd1e), and _that_ ironically requires _a_ version of the go compiler.

Finally, make sure the default Go directory is set in your `$PATH` so that any installed binaries will be available:

```bash
export PATH="$HOME/go/bin:$PATH"
```

## Python

The macOS comes only with Python 2.x and although the specific version _should_ (according to the Python docs) have the `pip` command available, that's not the case. So we have to install pip for Python2 manually using the _very old_ (but built-in) `easy_install` command:

```bash
sudo easy_install pip
```

Now when running `pip --version` we should see:

```
pip 19.0.3 from /Library/Python/2.7/site-packages/pip-19.0.3-py2.7.egg/pip (python 2.7)
```

At this point, in order to have a sane Python setup, we should look towards 'virtual environments'.

There are three aproaches we'll look at (each of them rely on [`pyenv`](https://github.com/pyenv/pyenv)):

1. [Pipenv](#pipenv) ([docs](https://pipenv.readthedocs.io/en/latest/install/))
2. [Poetry](#poetry) ([docs](https://poetry.eustace.io))
3. [pyenv-virtualenv](#pyenv-virtualenv) ([docs](https://github.com/pyenv/pyenv-virtualenv))

> Note: my preference is [pyenv-virtualenv](https://github.com/pyenv/pyenv-virtualenv) as it's simple and effective (read also my post: [Python Management and Project Dependencies](/posts/python-management/)).

We'll start by showing you how to install [`pipenv`](https://pipenv.readthedocs.io/en/latest/install/) which is a high-level abstraction across multiple tools (inc. [`pyenv`](https://github.com/pyenv/pyenv) and [`virtualenv`](https://virtualenv.readthedocs.io/)), then we'll move onto installing [Poetry](https://poetry.eustace.io). Lastly, we'll demonstrate `pyenv-virtualenv`.

### Pipenv

There is a brew install:

```bash
brew install pipenv
```

This will install Python 3.7.3 for you and so it'll be made available via `python3` and `pip3`.

Pipenv can't install Python versions for you. You'll need a tool such as `pyenv` which can be installed like so:

```bash
brew install pyenv
```

> Note: pyenv will also install `python-build` (no need to install that separately), but it's useful to know because the version of Python you want to install will be based on what's available from `python-build --definitions`.

So let's set-up a new Python environment (remember installing `pipenv` resulted in `python3` command being installed, and that's specifically version `3.7.3`, so we'll install a different Python3 version to that):

```bash
mkdir -p ~/Code/Python/3.7.1
cd ~/Code/Python/3.7.1

pyenv install 3.7.1

pipenv --python 3.7.1
pipenv install boto3 pytest structlog tornado
pipenv install --dev flake8 flake8-import-order mypy tox ipython

# notice the following command will fail as we haven't installed
# anything into the python3 version 3.7.3 that was installed when
# we installed pipenv...
ipython

# instead you can use pipenv's `run` subcommand to use Python 3.7.1
pipenv run python --version
pipenv run ipython

# pipenv's `shell` subcommand is an alternative to `pipenv run <command>`
# it'll drop you into a new shell which uses the relevant Python version
pipenv shell

# now these commands will work as they'll be using 3.7.1
python --version
ipython
```

If you don't have your `~/.bashrc` setup with `eval "$(pyenv init -)"`, then you won't have `/Users/integralist/.pyenv/shims` prefixed to your `$PATH` and so `pipenv` won't be able to locate your installed Python versions. 

If you don't want to modify your `$PATH` then you can manually specify the location of the Python interpreter version you want to use: 

```bash
pipenv --python /Users/integralist/.pyenv/versions/3.7.1/bin/python
```

> Note: virtual environments (and their `.project` files can be found here: `/Users/integralist/.local/share/virtualenvs`).

If you [have problems](https://github.com/pyenv/pyenv/wiki/Common-build-problems#build-failed-error-the-python-zlib-extension-was-not-compiled-missing-the-zlib) installing a Python version then you might need to reinstall XCode. The following is one solution that works for macOS Mojave:

```bash
xcode-select --install
sudo installer -pkg /Library/Developer/CommandLineTools/Packages/macOS_SDK_headers_for_macOS_10.14.pkg -target /
```

> Note: it's also useful to set-up autocompletion for pipenv in your `.bashrc` configuration file `eval "$(pipenv --completion)"`.

### Poetry

Poetry is better in that it's a cleaner and more isolated installation process (unlike Pipenv which requires us to `brew install python3`)

```bash
# install
curl -sSL https://raw.githubusercontent.com/sdispater/poetry/master/get-poetry.py | python

# reload .bash_profile and check poetry version
poetry --version

# update poetry to latest version
poetry self:update

# generate auto-complete for Homebrew installed version of bash
poetry completions bash > $(brew --prefix)/etc/bash_completion.d/poetry.bash-completion

# install python version
pyenv install 2.7.15

# check help for poetry init (which generates a `pyproject.toml`)
# poetry doesn't allow installing packages via cli (they need to be specified in toml)
poetry init -h

# create pyproject.toml interactively (see below for generated `pyproject.toml`)
# 
# notice [tool.poetry.dependencies] specifies the Python version used (this is required!).
poetry init

# install dependencies
poetry install

# add additional dependencies (use --dev for dev dependency)
poetry add requests <...>
poetry add --dev requests <...>

# execute commands within the virtual environment (e.g. dev dependency ipython was installed)
poetry run ipython

# load virtual environment permanently for the current shell (e.g. now python version will be the expected environment, not the OS version)
poetry shell
python --version

# here is a shortened Python3 example, as the above uses the OS default of Python2 for installing `2.7.15`
# where as if you tried to set the Python version in the `pyproject.toml` to `^3.7` it would fail as that Python version wouldn't be available
# it means whenever you want to setup a new Python3 environment, you'll need a compatible Python interpreter running first.
# e.g. if you want to install 3.7.1 you'll need 3.7.3 running first to execute Poetry (this isn't necessary with Python2 as we already had 2.7 available by the OS)
pyenv install 3.7.3
pyenv local 3.7.3
poetry add boto3 pytest structlog tornado
poetry add --dev flake8 flake8-import-order mypy tox ipython
```

Here is an example configuration file I use (`pyproject.toml`):

```toml
[tool.poetry]
name = "3.7.3"
version = "0.1.0"
description = ""
authors = ["Integralist"]

[tool.poetry.dependencies]
python = "^3.7"
boto3 = "^1.9"
pytest = "^4.4"
structlog = "^19.1"
tornado = "^6.0"

[tool.poetry.dev-dependencies]
black = { version = "*", allows-prereleases = true }
flake8 = "^3.7"
flake8-import-order = "^0.18.1"
mypy = "^0.701.0"
tox = "^3.9"
ipython = "^7.5"

[build-system]
requires = ["poetry>=0.12"]
build-backend = "poetry.masonry.api"
```

> Note: an older iteration of Black install was `black = { python=">3.6", version=">=19.3b0", allow_prereleases=true}` but `poetry check` would fail, so switched to updated version shown above.

### pyenv-virtualenv

This tool is a plugin for pyenv and is designed to manage virtual environments _only_, where as pipenv and poetry are toolkits designed for solving many different problems (one of which is virtual environments).

You install the plugin via Homebrew:

```
brew install pyenv-virtualenv
```

Next you add the following lines to your `~/.bashrc`:

```bash
eval "$(pyenv init -)"
eval "$(pyenv virtualenv-init -)"
```

Now within some directory you can define a new virtual environment using:

```bash
pyenv virtualenv 3.7.1 testing-plugin-with-3.7.1
```

To see the available virtual environments:

```bash
$ pyenv virtualenvs

3.7.1/envs/testing-plugin-with-3.7.1 (created from /Users/integralist/.pyenv/versions/3.7.1)
testing-plugin-with-3.7.1 (created from /Users/integralist/.pyenv/versions/3.7.1)
```

To activate and deactivate the virtual environment:

```bash
pyenv activate testing-plugin-with-3.7.1
pyenv deactivate
```

Simple.

### Python Packages

Here are some packages I like to install as a general rule...

- [black](https://github.com/python/black): formatter (like golang's `gofmt`)
- [flake8](http://flake8.pycqa.org): linter
- [flake8-import-order](https://github.com/PyCQA/flake8-import-order): validates imports
- [mypy](http://www.mypy-lang.org): static analysis
- [ipython](https://ipython.org): REPL
- [pytest](https://pytest.org): testing framework
- [structlog](http://www.structlog.org/en/stable/): structured logging
- [tornado](https://www.tornadoweb.org): async web framework
- [boto3](https://boto3.amazonaws.com/v1/documentation/api/latest/index.html): AWS CLI tool
- [tox](https://tox.readthedocs.io/en/latest/): generic virtualenv management and testing tool

> Note: for an example of how to configure Flake8 and its plugins, see [this gist](https://gist.github.com/0ce27db1d7294f3af9896c0807ccfeed).

I would also strongly recommend using `pipx` and installing programs from there, such as `isort` (which I then reference in my `.vimrc`).

## Vim

You can either install more recent version of vim via Homebrew:

```bash
brew install vim
```

Or you can manually compile vim yourself:

> Note: I manually compile vim as I need Python3 support baked in, which Homebrew's version no longer does (it used to, but not any more). Python3 support means my Python linting tools will work as expected.

```bash
git clone https://github.com/vim/vim.git

cd vim

./configure --with-features=huge \
            --enable-multibyte \
            --enable-rubyinterp=yes \
            --enable-python3interp=yes \  # relies on `brew install python3`
            --enable-perlinterp=yes \
            --enable-luainterp=yes \
            --enable-gui=gtk2 \
            --enable-cscope \
            --prefix=/usr/local
            
make && make install
```

The above code will copy the compiled vim binary into `/usr/local/bin` so `which vim` will show `/usr/local/bin/vim`

> Note: you can also specify the path to the Python 3 interpreter if you don't want to rely on a global Python3 interpreter (`--with-python3-config-dir`). This gets more confusing when you have packages such as [Black](https://github.com/psf/black) that needs to be installed to that global interpreter.

Configure vim with [vim-plug](https://github.com/junegunn/vim-plug) plugin manager:

```bash
curl -fLo ~/.vim/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
curl -LSso ~/.vimrc https://raw.githubusercontent.com/Integralist/dotfiles/master/.vimrc
```

Ensure Vim is configured with spell checking options:

```bash
vim -E -s <<EOF
:set spell
:quit
EOF
```

Install plugins by opening vim and executing:

```
:PlugInstall
```

> Note: [fzf](https://github.com/junegunn/fzf) doesn't need a brew install when installed via vim. See my `.vimrc` configuration file for more details, but in essence it contains: `Plug 'junegunn/fzf', { 'dir': '~/.fzf', 'do': './install --all' }`

Also ensure the Golang environment has what it needs by executing:

```
:GoInstallBinaries
```

## Tmux

Install `tmux`:

```bash
brew install tmux
```

Configure tmux and expose `tmuxy` command (defined in my `~/.bashrc` for quickly spinning up a new working environment):

```bash
curl -LSso ~/.tmux.conf https://raw.githubusercontent.com/Integralist/dotfiles/master/.tmux.conf
curl -LSso ~/tmux.sh https://raw.githubusercontent.com/Integralist/dotfiles/master/tmux.sh
```

> Note: check `$PATH` to make sure tmux isn't double setting values in your PATH as it starts up. If it does you can check an older version of my [`~/.bash_profile`](https://github.com/Integralist/dotfiles/blob/cc906bd14636543e71d9c034d6507f5986a80bbd/.bash_profile#L18-L21) for a work-around.

## Miscellaneous

Not every app can be installed via Homebrew. [Monosnap](https://monosnap.com/welcome) is one such example.

If you want an easy way to hide menu bar items, then try the [hidden-bar](https://apps.apple.com/app/hidden-bar/id1452453066) app ([github](https://github.com/dwarvesf/hidden))

Also, if you're into torrents, then the transmission server/client (`npm install -g t-get`) might be of interest to you.

## macOS

It can be cool to configure the macOS via the terminal, things like mouse cursor speed or keyboard repeat key performance. But unfortunately that all changed with macOS Mojave and I couldn't be bothered to figure out the correct way to do it via the terminal when doing the setup via the GUI is just as quick (and I know the few things I like to tweak off-by-heart).

For example, you used to be able to do things like:

```bash
defaults write NSGlobalDomain ApplePressAndHoldEnabled -bool false
```

But since macOS Mojave those settings and namespaces seem to have changed. If you're interested in figuring it out, then I'd recommend starting with:

```bash
defaults read
```

The above command will display _all_ the current macOS settings for you. From there you can drill down into individual namespaces like so:

```bash
defaults read "Apple Global Domain" com.apple.mouse.tapBehavior
```

> Note: [here](https://github.com/Integralist/dotfiles/blob/cc906bd14636543e71d9c034d6507f5986a80bbd/bootstrap.sh#L7-L53) are the settings I used to configure via the terminal.

One thing I like to do is to make sure macOS' "Spaces" feature doesn't rearrange spaces based on their recent usage, and to do that you need to open up the 'Mission Control' settings panel and disable the option:

```
Automatically rearrange Spaces based on most recent use
```

## Homebrew Output

Finally, for those interested, below is the output of installing Homebrew for the first time. I like to see what Homebrew creates, so in future if I ever want to know where something should exist I can refer back to this as a reference point:


```
Integralist-MBP:~ integralist /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
==> This script will install:
/usr/local/bin/brew
/usr/local/share/doc/homebrew
/usr/local/share/man/man1/brew.1
/usr/local/share/zsh/site-functions/_brew
/usr/local/etc/bash_completion.d/brew
/usr/local/Homebrew
==> The following existing directories will be made group writable:
/usr/local/bin
==> The following existing directories will have their owner set to integralist:
/usr/local/bin
==> The following existing directories will have their group set to admin:
/usr/local/bin
==> The following new directories will be created:
/usr/local/etc
/usr/local/include
/usr/local/lib
/usr/local/sbin
/usr/local/share
/usr/local/var
/usr/local/opt
/usr/local/share/zsh
/usr/local/share/zsh/site-functions
/usr/local/var/homebrew
/usr/local/var/homebrew/linked
/usr/local/Cellar
/usr/local/Caskroom
/usr/local/Homebrew
/usr/local/Frameworks
==> The Xcode Command Line Tools will be installed.

Press RETURN to continue or any other key to abort
==> /usr/bin/sudo /bin/chmod u+rwx /usr/local/bin
Password:
==> /usr/bin/sudo /bin/chmod g+rwx /usr/local/bin
==> /usr/bin/sudo /usr/sbin/chown integralist /usr/local/bin
==> /usr/bin/sudo /usr/bin/chgrp admin /usr/local/bin
==> /usr/bin/sudo /bin/mkdir -p /usr/local/etc /usr/local/include /usr/local/lib /usr/local/sbin /usr/local/share /usr/local/var /usr/local/opt /usr/local/share/zsh /usr/local/share/zsh/site-functions /usr/local/var/homebrew /usr/local/var/homebrew/linked /usr/local/Cellar /usr/local/Caskroom /usr/local/Homebrew /usr/local/Frameworks
==> /usr/bin/sudo /bin/chmod g+rwx /usr/local/etc /usr/local/include /usr/local/lib /usr/local/sbin /usr/local/share /usr/local/var /usr/local/opt /usr/local/share/zsh /usr/local/share/zsh/site-functions /usr/local/var/homebrew /usr/local/var/homebrew/linked /usr/local/Cellar /usr/local/Caskroom /usr/local/Homebrew /usr/local/Frameworks
==> /usr/bin/sudo /bin/chmod 755 /usr/local/share/zsh /usr/local/share/zsh/site-functions
==> /usr/bin/sudo /usr/sbin/chown integralist /usr/local/etc /usr/local/include /usr/local/lib /usr/local/sbin /usr/local/share /usr/local/var /usr/local/opt /usr/local/share/zsh /usr/local/share/zsh/site-functions /usr/local/var/homebrew /usr/local/var/homebrew/linked /usr/local/Cellar /usr/local/Caskroom /usr/local/Homebrew /usr/local/Frameworks
==> /usr/bin/sudo /usr/bin/chgrp admin /usr/local/etc /usr/local/include /usr/local/lib /usr/local/sbin /usr/local/share /usr/local/var /usr/local/opt /usr/local/share/zsh /usr/local/share/zsh/site-functions /usr/local/var/homebrew /usr/local/var/homebrew/linked /usr/local/Cellar /usr/local/Caskroom /usr/local/Homebrew /usr/local/Frameworks
==> /usr/bin/sudo /bin/mkdir -p /Users/integralist/Library/Caches/Homebrew
==> /usr/bin/sudo /bin/chmod g+rwx /Users/integralist/Library/Caches/Homebrew
==> /usr/bin/sudo /usr/sbin/chown integralist /Users/integralist/Library/Caches/Homebrew
==> Searching online for the Command Line Tools
==> /usr/bin/sudo /usr/bin/touch /tmp/.com.apple.dt.CommandLineTools.installondemand.in-progress
==> Installing Command Line Tools (macOS Mojave version 10.14) for Xcode-10.2
==> /usr/bin/sudo /usr/sbin/softwareupdate -i Command\ Line\ Tools\ (macOS\ Mojave\ version\ 10.14)\ for\ Xcode-10.2
Software Update Tool


Downloading Command Line Tools (macOS Mojave version 10.14) for Xcode
Downloaded Command Line Tools (macOS Mojave version 10.14) for Xcode
Installing Command Line Tools (macOS Mojave version 10.14) for Xcode
Done with Command Line Tools (macOS Mojave version 10.14) for Xcode
Done.
==> /usr/bin/sudo /bin/rm -f /tmp/.com.apple.dt.CommandLineTools.installondemand.in-progress
==> /usr/bin/sudo /usr/bin/xcode-select --switch /Library/Developer/CommandLineTools
==> Downloading and installing Homebrew...
remote: Enumerating objects: 63, done.
remote: Counting objects: 100% (63/63), done.
remote: Compressing objects: 100% (46/46), done.
remote: Total 121248 (delta 31), reused 41 (delta 15), pack-reused 121185
Receiving objects: 100% (121248/121248), 28.67 MiB | 18.10 MiB/s, done.
Resolving deltas: 100% (88696/88696), done.
From https://github.com/Homebrew/brew
 * [new branch]      master     -> origin/master
 * [new tag]         0.1        -> 0.1
 * [new tag]         0.2        -> 0.2
 * [new tag]         0.3        -> 0.3
 * [new tag]         0.4        -> 0.4
 * [new tag]         0.5        -> 0.5
 * [new tag]         0.6        -> 0.6
 * [new tag]         0.7        -> 0.7
 * [new tag]         0.7.1      -> 0.7.1
 * [new tag]         0.8        -> 0.8
 * [new tag]         0.8.1      -> 0.8.1
 * [new tag]         0.9        -> 0.9
 * [new tag]         0.9.1      -> 0.9.1
 * [new tag]         0.9.2      -> 0.9.2
 * [new tag]         0.9.3      -> 0.9.3
 * [new tag]         0.9.4      -> 0.9.4
 * [new tag]         0.9.5      -> 0.9.5
 * [new tag]         0.9.8      -> 0.9.8
 * [new tag]         0.9.9      -> 0.9.9
 * [new tag]         1.0.0      -> 1.0.0
 * [new tag]         1.0.1      -> 1.0.1
 * [new tag]         1.0.2      -> 1.0.2
 * [new tag]         1.0.3      -> 1.0.3
 * [new tag]         1.0.4      -> 1.0.4
 * [new tag]         1.0.5      -> 1.0.5
 * [new tag]         1.0.6      -> 1.0.6
 * [new tag]         1.0.7      -> 1.0.7
 * [new tag]         1.0.8      -> 1.0.8
 * [new tag]         1.0.9      -> 1.0.9
 * [new tag]         1.1.0      -> 1.1.0
 * [new tag]         1.1.1      -> 1.1.1
 * [new tag]         1.1.10     -> 1.1.10
 * [new tag]         1.1.11     -> 1.1.11
 * [new tag]         1.1.12     -> 1.1.12
 * [new tag]         1.1.13     -> 1.1.13
 * [new tag]         1.1.2      -> 1.1.2
 * [new tag]         1.1.3      -> 1.1.3
 * [new tag]         1.1.4      -> 1.1.4
 * [new tag]         1.1.5      -> 1.1.5
 * [new tag]         1.1.6      -> 1.1.6
 * [new tag]         1.1.7      -> 1.1.7
 * [new tag]         1.1.8      -> 1.1.8
 * [new tag]         1.1.9      -> 1.1.9
 * [new tag]         1.2.0      -> 1.2.0
 * [new tag]         1.2.1      -> 1.2.1
 * [new tag]         1.2.2      -> 1.2.2
 * [new tag]         1.2.3      -> 1.2.3
 * [new tag]         1.2.4      -> 1.2.4
 * [new tag]         1.2.5      -> 1.2.5
 * [new tag]         1.2.6      -> 1.2.6
 * [new tag]         1.3.0      -> 1.3.0
 * [new tag]         1.3.1      -> 1.3.1
 * [new tag]         1.3.2      -> 1.3.2
 * [new tag]         1.3.3      -> 1.3.3
 * [new tag]         1.3.4      -> 1.3.4
 * [new tag]         1.3.5      -> 1.3.5
 * [new tag]         1.3.6      -> 1.3.6
 * [new tag]         1.3.7      -> 1.3.7
 * [new tag]         1.3.8      -> 1.3.8
 * [new tag]         1.3.9      -> 1.3.9
 * [new tag]         1.4.0      -> 1.4.0
 * [new tag]         1.4.1      -> 1.4.1
 * [new tag]         1.4.2      -> 1.4.2
 * [new tag]         1.4.3      -> 1.4.3
 * [new tag]         1.5.0      -> 1.5.0
 * [new tag]         1.5.1      -> 1.5.1
 * [new tag]         1.5.10     -> 1.5.10
 * [new tag]         1.5.11     -> 1.5.11
 * [new tag]         1.5.12     -> 1.5.12
 * [new tag]         1.5.13     -> 1.5.13
 * [new tag]         1.5.14     -> 1.5.14
 * [new tag]         1.5.2      -> 1.5.2
 * [new tag]         1.5.3      -> 1.5.3
 * [new tag]         1.5.4      -> 1.5.4
 * [new tag]         1.5.5      -> 1.5.5
 * [new tag]         1.5.6      -> 1.5.6
 * [new tag]         1.5.7      -> 1.5.7
 * [new tag]         1.5.8      -> 1.5.8
 * [new tag]         1.5.9      -> 1.5.9
 * [new tag]         1.6.0      -> 1.6.0
 * [new tag]         1.6.1      -> 1.6.1
 * [new tag]         1.6.10     -> 1.6.10
 * [new tag]         1.6.11     -> 1.6.11
 * [new tag]         1.6.12     -> 1.6.12
 * [new tag]         1.6.13     -> 1.6.13
 * [new tag]         1.6.14     -> 1.6.14
 * [new tag]         1.6.15     -> 1.6.15
 * [new tag]         1.6.16     -> 1.6.16
 * [new tag]         1.6.17     -> 1.6.17
 * [new tag]         1.6.2      -> 1.6.2
 * [new tag]         1.6.3      -> 1.6.3
 * [new tag]         1.6.4      -> 1.6.4
 * [new tag]         1.6.5      -> 1.6.5
 * [new tag]         1.6.6      -> 1.6.6
 * [new tag]         1.6.7      -> 1.6.7
 * [new tag]         1.6.8      -> 1.6.8
 * [new tag]         1.6.9      -> 1.6.9
 * [new tag]         1.7.0      -> 1.7.0
 * [new tag]         1.7.1      -> 1.7.1
 * [new tag]         1.7.2      -> 1.7.2
 * [new tag]         1.7.3      -> 1.7.3
 * [new tag]         1.7.4      -> 1.7.4
 * [new tag]         1.7.5      -> 1.7.5
 * [new tag]         1.7.6      -> 1.7.6
 * [new tag]         1.7.7      -> 1.7.7
 * [new tag]         1.8.0      -> 1.8.0
 * [new tag]         1.8.1      -> 1.8.1
 * [new tag]         1.8.2      -> 1.8.2
 * [new tag]         1.8.3      -> 1.8.3
 * [new tag]         1.8.4      -> 1.8.4
 * [new tag]         1.8.5      -> 1.8.5
 * [new tag]         1.8.6      -> 1.8.6
 * [new tag]         1.9.0      -> 1.9.0
 * [new tag]         1.9.1      -> 1.9.1
 * [new tag]         1.9.2      -> 1.9.2
 * [new tag]         1.9.3      -> 1.9.3
 * [new tag]         2.0.0      -> 2.0.0
 * [new tag]         2.0.1      -> 2.0.1
 * [new tag]         2.0.2      -> 2.0.2
 * [new tag]         2.0.3      -> 2.0.3
 * [new tag]         2.0.4      -> 2.0.4
 * [new tag]         2.0.5      -> 2.0.5
 * [new tag]         2.0.6      -> 2.0.6
 * [new tag]         2.1.0      -> 2.1.0
HEAD is now at 1c655916f Merge pull request #5993 from amyspark/drop-unzip-in-macos
==> Homebrew is run entirely by unpaid volunteers. Please consider donating:
  https://github.com/Homebrew/brew#donations
==> Tapping homebrew/core
Cloning into '/usr/local/Homebrew/Library/Taps/homebrew/homebrew-core'...
remote: Enumerating objects: 4958, done.
remote: Counting objects: 100% (4958/4958), done.
remote: Compressing objects: 100% (4729/4729), done.
remote: Total 4958 (delta 51), reused 1272 (delta 38), pack-reused 0
Receiving objects: 100% (4958/4958), 3.98 MiB | 7.98 MiB/s, done.
Resolving deltas: 100% (51/51), done.
Tapped 2 commands and 4743 formulae (5,000 files, 12.4MB).
Already up-to-date.
==> Installation successful!

==> Homebrew has enabled anonymous aggregate formulae and cask analytics.
Read the analytics documentation (and how to opt-out) here:
  https://docs.brew.sh/Analytics

==> Homebrew is run entirely by unpaid volunteers. Please consider donating:
  https://github.com/Homebrew/brew#donations
==> Next steps:
- Run `brew help` to get started
- Further documentation: 
    https://docs.brew.sh
```
