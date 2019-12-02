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

- [Introduction](#introduction)
- [Virtual Environments](#virtual-environments)
- [Installing Python Versions](#installing-python-versions)
- [Setting Up Virtual Environments](#setting-up-virtual-environments)
- [Only Virtual Environments](#only-virtual-environments)
- [Managing Dependencies](#managing-dependencies)

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

## Managing Dependencies

When it comes to dealing with specific dependency versions, I like to use the method [Kenneth Reitz](https://www.kennethreitz.org/essays/a-better-pip-workflow) published back in 2016.

> **Note**: this method keeps with the traditional `requirements.txt` file as utilized by [Pip](https://pip.pypa.io/en/stable/). I mention this as you'll notice with other tools (such as Pipenv or Poetry), that they move away from this established format and that can be a bit disruptive in terms of how Python teams have traditionally worked. I'm not saying it's a bad thing, but change isn't always good.

### Problem Summary

Here is a summary of the problem we're trying to solve:

The `requirements.txt` file typically doesn't include the sub-dependencies required by your top-level dependencies (because that would require you to manually identify them and to type them all out, something that should be an automated process and so in practice never happens manually). 

e.g. you specify a top-level dependency of `foo` (which might install version 1.0), but that internally requires the use of other third-party packages such as `bar` and `baz` (and specific versions for each of them).

But a `pip install` from a file that only includes the top-level dependencies could (over time) result in different sub-dependency versions being installed by either different members of your team or via your deployment platform.

e.g. if you don't specify a version for `foo`, then in a months time when someone else (or your deployment platform) runs `pip install` it will attempt to install the latest version of `foo` which might be version 2.0 (and subsequently the third-party packages it uses might also change).

To avoid that people have come to use `pip freeze` after doing a `pip install` to overwrite their `requirements.txt` with a list of _all_ dependencies and their explicit versions. 

This solves the issue of installing from `requirements.txt` in a months time when lots of your top-level dependencies release new breaking versions. 

The problem now is that you have to manually search for the top-level dependencies (in this new larger/more-indepth `requirements.txt` file) and update them manually. Doing this might break things as you now don't know what the sub-dependency versions should be set to.

### Solution

So the approach we take with any project is to define a `requirements-to-freeze.txt` file. This file will contain all your project's top-level dependencies (inc. any _explicit_ versions required), for example:

```
requests[security]
flask
gunicorn==19.4.5
```

Next we can generate our actual `requirements.txt` file based upon the contents of `requirements-to-freeze.txt` using the `pip freeze` command, like so:

```
python -m pip install -r requirements-to-freeze.txt
python -m pip freeze > requirements.txt
```

Which will result in a `requirements.txt` file that looks something like:

```
cffi==1.5.2
cryptography==1.2.2
enum34==1.1.2
Flask==0.10.1
gunicorn==19.4.5
idna==2.0
ipaddress==1.0.16
itsdangerous==0.24
Jinja2==2.8
MarkupSafe==0.23
ndg-httpsclient==0.4.0
pyasn1==0.1.9
pycparser==2.14
pyOpenSSL==0.15.1
requests==2.9.1
six==1.10.0
Werkzeug==0.11.4
```

This means you'll never manually update `requirements.txt` again. Any time you need to update a dependency you'll do it in `requirements-to-freeze.txt`, then re-run:

```
python -m pip install -r requirements-to-freeze.txt
python -m pip freeze > requirements.txt
```

Or instead of manually updating the dependencies in `requirements-to-freeze.txt` you could use the `--upgrade` flag:

```
python -m pip install -r requirements-to-freeze.txt --upgrade
python -m pip freeze > requirements.txt
```
