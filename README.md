# graphql.natwelch.com

A new backend for graphql.natwelch.com.

[![Build Status](https://travis-ci.org/icco/graphql.svg?branch=master)](https://travis-ci.org/icco/graphql)

Docs: [godoc.org/github.com/icco/graphql](https://godoc.org/github.com/icco/graphql)

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

This site is hosted at <https://graphql.natwelch.com>. It runs out of a docker container on Google Kubernetes. It has a postgres backend. This started as a rewrite of a previous project, natnatnat. Its [readme](https://github.com/icco/natnatnat/blob/master/README.md) walks through a lot of the previous inspiration.

## Documentation

You can explore this api by looking at [schema.graphql]() and reading the descriptions. See https://facebook.github.io/graphql/June2018/#sec-Descriptions for an explanation of the description schema.
