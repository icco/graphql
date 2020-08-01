# graphql.natwelch.com

A new backend for graphql.natwelch.com.

[![Build Status](https://travis-ci.org/icco/graphql.svg?branch=master)](https://travis-ci.org/icco/graphql) [![Go Report Card](https://goreportcard.com/badge/github.com/icco/graphql)](https://goreportcard.com/report/github.com/icco/graphql)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ficco%2Fgraphql.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Ficco%2Fgraphql?ref=badge_shield)

Docs: [godoc.org/github.com/icco/graphql](https://godoc.org/github.com/icco/graphql)

The next iteration in Nat's content management system. Previous versions include:

 * [tumble.io](http://github.com/icco/tumble)
 * [pseudoweb.net](http://github.com/icco/pseudoweb)
 * [natnatnat](http://github.com/icco/natnatnat)


## Install

This repo requires Go 1.11 to be installed.

 1. Start postgres on your local machine with a database called writing.
 2. Copy `local.env` to `.env`
 3. `env $(cat .env) go run -v ./server` to start the server.
 4. Visit <http://localhost:8080/> which has a default graphql client.

### Example Env

```
DATABASE_URL=postgres://localhost/writing?sslmode=disable&binary_parameters=yes
SESSION_SECRET="random string"
OAUTH2_CLIENTID=something.apps.googleusercontent.com
OAUTH2_SECRET=1234567890
OAUTH2_REDIRECT=http://localhost:8080/callback
PORT=9393
```

### Auth

This uses Auth0 to generate logins. To save yourself setting up the Auth0, you can generate an API key for testing by creating a user. To create a user for testing, run the following insert SQL:

```sql
INSERT INTO users (id, role, created_at, modified_at) VALUES ('test', 'admin', now(), now());
```

Then get your API key:

```sql
SELECT apikey from users where id = 'test';
```

And then set that as the value of the `X-API-AUTH` on all of your requests to graphql.

## Design

This site is hosted at <https://graphql.natwelch.com>. It runs out of a docker container on Google Kubernetes. It has a postgres backend. This started as a rewrite of a previous project, natnatnat. Its [readme](https://github.com/icco/natnatnat/blob/master/README.md) walks through a lot of the previous inspiration.

We use <https://github.com/99designs/gqlgen> to generate a lot of the files. If you modify [schema.graphql](), please run `go run ./scripts/gqlgen.go -v` to update things.

## Documentation

You can explore this api by looking at [schema.graphql]() and reading the descriptions. See <https://facebook.github.io/graphql/June2018/#sec-Descriptions> for an explanation of the description schema.


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ficco%2Fgraphql.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Ficco%2Fgraphql?ref=badge_large)
