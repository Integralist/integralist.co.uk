---
title: "Python Management"
date: 2019-12-01T11:25:19Z
categories:
  - "code"
  - "development"
  - "guide"
tags:
  - "management"
  - "installation"
  - "python"
draft: false
---

## Introduction

This blog post aims to demonstrate the most practical way to install multiple versions of Python, and of setting up 'virtual environments' for macOS users. 

Why is this not a simple thing to do? I hear you ask! Well, because we don't live in a perfect world and sometimes the overhead of 'convenience' can be worth the cost otherwise incurred.

I'll start by describing briefly what a virtual environment is, and then I'll move onto the point of this article and the options we have available to us.

## Virtual Environments

> A cooperatively isolated runtime environment that allows Python users and applications to install and upgrade Python distribution packages without interfering with the behaviour of other Python applications running on the same system. -- [Python Glossary](https://docs.python.org/3/glossary.html#term-virtual-environment)

To simplify: if you work on multiple Python projects and each project uses the same external dependency (e.g. the [Requests](https://requests.readthedocs.io/en/master/) HTTP library), then it's possible your application (for _each_ project) will be written to support a specific version of that dependency.

Having a virtual environment setup for each project means you can have project specific Python package installs. For example, project 'foo' can use the Request library version 1.0 while project 'bar' can use version 2.0 (which might introduce a different API).

> Note: the official documentation on Virtual Environments can be found [here](https://docs.python.org/3/tutorial/venv.html).

## Installing Python Versions

So there are in fact _two_ ways of installing Python versions, but it's much less practical in reality and so i'm only going to demonstrate _one_ of them.

The two ways are:

- Manual compilation
- External build tool

As you can imagine, manually compiling Python would be the more flexible solution, but the harsh reality is that compiling Python requires very specific dependencies that can be hard to get right (and most often it goes wrong).

So instead we'll focus on the latter: an external tool called `pyenv` which internally uses another external tool called `python-build`.

- `pyenv`: let's us switch Python versions easily
- `python-build` let's us install Python versions easily.

You don't need to install `python-build` directly as it'll be installed when you install `pyenv`.

To install `pyenv` execute:

```
brew install pyenv
```

> Note: yes, this means you need to be using [Homebrew](https://brew.sh/), but let's face it, it's the defacto standard for macOS package management.

Once installed you'll be able to use the following commands:

- `python-build --definitions`: list all versions of Python available to be installed
- `pyenv install <version>`: install the version of Python you need

Once you have installed the version of Python you need, now you just need to remember to 'activate' it whenever you're working on your project that requires that specific version of Python. To do that, you'll need to do two things (and you only need to do them once):

- add `eval "$(pyenv init -)"` to your `.bashrc` (or shell of choice)
- execute `pyenv local <version>` in your project directory

What the first point does is it'll allow your shell to respond to any `.python-version` file found within a directory on your computer. This file will contain a Python version.

What generates the `.python-version` file is the latter point.

## Setting Up Virtual Environments

To setup virtual environments with Python is actually very simple (see [next section](#only-virtual-environments)), but not compatible when using an external build tool such as `pyenv` because of where `pyenv` installs Python binaries and how it switches between versions.

But luckily there is an extension to `pyenv` called `pyenv-virtualenv` which can be installed like so:

```
brew install pyenv-virtualenv
```

Once installed, setting up a new virtual environment is as simple as:

```
pyenv virtualenv foobar
pyenv activate foobar
```

## Only Virtual Environments

If you installed Python using Homebrew (e.g. `brew install python3`) and you don't care what version of Python you have, you only care about having virtual environments, then you can utilize the built-in `venv` module like so:

```
python3 -m venv /foo/bar
source /foo/bar/bin/activate
python3 -m pip install <dependencies>
```

This will ensure you only install third-party packages/modules into the specific virtual environment you've just activated.

The downside of this very simple approach is that installing Python via Homebrew means you'll have only a single version of Python installed and you have no control over what that version is, let alone have multiple versions installed (as the `python3` command will be overwritten each time).
