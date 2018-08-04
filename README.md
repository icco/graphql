# writing-be

A new backend for writing.natwelch.com.

[![Build Status](https://travis-ci.org/icco/writing.svg?branch=master)](https://travis-ci.org/icco/writing)

Docs: [godoc.org/github.com/icco/writing](https://godoc.org/github.com/icco/writing)

The next iteration in Nat's content management system. Previous versions include:

 * [tumble.io](http://github.com/icco/tumble)
 * [pseudoweb.net](http://github.com/icco/pseudoweb)
 * [natnatnat](http://github.com/icco/natnatnat)


## Install

These directions are for OSX and assume you have [homebrew](http://brew.sh/) installed.

 1. Run `brew bundle`
 2. Run `make update`
 3. Run `make` to run locally

## Design

This site is hosted at <http://writing.natwelch.com>. It runs out of a docker container on Google Kubernetes. It has a postgres backend. This started as a rewrite of a previous project, natnatnat. Its [readme](https://github.com/icco/natnatnat/blob/master/README.md) walks through a lot of the previous inspiration.
