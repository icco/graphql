# graphql.natwelch.com

A new backend for graphql.natwelch.com.

[![Build Status](https://travis-ci.org/icco/graphql.svg?branch=master)](https://travis-ci.org/icco/graphql)

Docs: [godoc.org/github.com/icco/graphql](https://godoc.org/github.com/icco/graphql)

The next iteration in Nat's content management system. Previous versions include:

 * [tumble.io](http://github.com/icco/tumble)
 * [pseudoweb.net](http://github.com/icco/pseudoweb)
 * [natnatnat](http://github.com/icco/natnatnat)


## Install

This repo requires Go 1.11 to be installed.

 1. Start postgres on your local machine with a database called writing.
 2. Copy `local.env` to `.env`
 3. If you're going to be playing with auth, Follow https://support.google.com/googleapi/answer/6158849?hl=en and create a "Web Application" OAuth2.0 config
 4. `env $(cat .env) go run -v ./server` to start the server.

## Design

This site is hosted at <https://graphql.natwelch.com>. It runs out of a docker container on Google Kubernetes. It has a postgres backend. This started as a rewrite of a previous project, natnatnat. Its [readme](https://github.com/icco/natnatnat/blob/master/README.md) walks through a lot of the previous inspiration.

## Documentation

You can explore this api by looking at [schema.graphql]() and reading the descriptions. See https://facebook.github.io/graphql/June2018/#sec-Descriptions for an explanation of the description schema.
