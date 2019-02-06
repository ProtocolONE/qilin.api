[![Build Status](https://travis-ci.org/ProtocolONE/qilin.api.svg?branch=master)](https://travis-ci.org/ProtocolONE/qilin.api) [![codecov](https://codecov.io/gh/ProtocolONE/qilin.api/branch/master/graph/badge.svg)](https://codecov.io/gh/ProtocolONE/qilin.api)[![Go Report Card](https://goreportcard.com/badge/github.com/ProtocolONE/qilin.api)](https://goreportcard.com/report/github.com/ProtocolONE/qilin.api)

# Qilin Management API

Qilin is an open source tool facilitating creation, distribution and activation of licenses for game content. 

The mission of Qilin n is to enable developers, publishers, platforms and stores to distribute games,
reducing their efforts on concluding and drafting contracts and exchanging documentation to a
minimum, and providing them with comprehensive statistics in real time.

Our solution is a part of Protocol One IAAS, it is actively used in Storefront constructor. Qilin
can be installed as a component on an existing P1-independent system, employing its own
hardware or cloud platform.

## Get started

Qilin management API designed to be launched with Kubernetes and handle all configuration from env variables:

| Variable                      | Default                                                           | Description                                                                                                                                |
|-------------------------------|-------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| QILIN_API_SERVER_PORT              | 8080                                                              | HTTP port to listed API requests.                                                                                                          |
| QILIN_API_SERVER_ALLOW_ORIGINS     | *                                                                 | Comma separated list of [CORS domains](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin).             |
| QILIN_API_SERVER_ALLOW_CREDENTIALS | false                                                             | Look at [CORS documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials) about this value. |
| QILIN_API_SERVER_DEBUG             | false                                                             | Enable debug mode for [echo based](https://echo.labstack.com/) server.                                                                     |
| QILIN_API_DATABASE_DSL             | See below                                                         | [GORM Postgres DSL](http://doc.gorm.io/database.html#connecting-to-a-database) string                                                      |
| QILIN_API_DATABASE_DEBUG           | false                                                             | Enable logmode for Postgress.                                                                                                              |
| QILIN_API_JWT_SECRET               |                                                                   | BASE64 encoded JWT secret key. Here is no default value, it should be provided.                                                            |
| QILIN_API_JWT_ALGORITHM            | HS256                                                             | JWT Algorithm.                                                                                                                             |
| QILIN_API_LOG_LEVEL                | debug                                                             | Default logging level in application.                                                                                                      |
| QILIN_API_LOG_REPORT_CALLER        | false                                                             | Loggin stack trace enable.                                                                                                                 |

Default database DSL is `postgres://postgres:postgres@localhost:5432/qilin?sslmode=disable`. 

This version of server use dump mail sender in current implementation. You may also configure it with env variables


| Variable                | Default   | Description                                                             |
|-------------------------|-----------|-------------------------------------------------------------------------|
| QILIN_API_MAILER_HOST        | localhost | Email server host.                                                      |
| QILIN_API_MAILER_PORT        | 25        | Email server port.                                                      |
| QILIN_API_MAILER_USERNAME    |           | Email server username. Here is no default value, it should be provided. |
| QILIN_API_MAILER_PASSWORD    |           | Email server password. Here is no default value, it should be provided. |
| QILIN_API_MAILER_REPLY_TO    |           | Reply-to value. Here is no default value, it may be provided.           |
| QILIN_API_MAILER_FROM        |           | From value. Here is no default value, it may be provided.               |
| QILIN_API_MAILER_SKIP_VERIFY | true      | Skip validate TLS on mail server connection.                            |
 
## Features

 * Sales Growth. ​Individual pricing, cross-platform and cross-shop data create an
individual feature list and price offers for each user, increasing sales conversions and
their quantity.
 * New Games. ​Using e-sign in the Qilin agreement is all you have to do to access the
listing of games within the Protocol One ecosystem.
 * All-platform release.​ Games and developers using Qilin can have a release on all
platforms and stores within the ecosystem stores with a single click.
 * Quick withdrawal of funds.​ Qilin can credit the end stores — now you can withdraw
funds whenever you wish, ignoring payment conditions of the end stores (minimum
amount threshold, monthly or quarterly payments).
 * Settings. ​Real time regional price changes and planning the discounts on all
platforms and stores without restrictions.
 * Statistics.​ Timely information on the amount of keys sold, the geography of sales,
final prices and VAT updated in real time.
 * Key Streaming​. Create keys in Qinin without restrictions, in real time, for every
purchase. For the distribution of 3rd party keys, a single set of keys is used to
distribute to any number of end platforms.
2
 * Advanced regional restrictions system.​ A key coming from a “cheaper” region can
either be prevented from activation, or the user can be offered to pay the difference.
 * Royalty reports.​ Embedded legally verifiable royalty reports for an arbitrary period of
time, including cumulative figures across all platforms and stores.
 * Unified SDK.​ Qilin SDK used in the game can achieve unified integration with
Steam, Gog, Kartridge and other platforms to manage authorization, achievements,
payments and cloud saves. A single build can be used across all platforms

## Supported go versions
We support the major Go versions, which are 1.11 at the moment.

## Contributing
Please feel free to submit issues, fork the repository and send pull requests!

When submitting an issue, we ask that you please include a complete test function that demonstrates the issue. Extra credit for those using Testify to write the test code that demonstrates it.
