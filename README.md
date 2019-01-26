<!--
title: TODO
description: This example demonstrates how to setup a simple HTTP endpoint in Go.
layout: Doc
framework: v1
platform: AWS
language: Go
authorLink: 'https://github.com/sebito91'
authorName: 'Sebastian Borza'
authorAvatar: 'https://avatars0.githubusercontent.com/u/3159454?v=4&s=140'
-->

## Endpoints
---
**validateCode** - checks to see whether a coupon code or instagram account name is valid
| Verb | Endpoint |
| ----------- | ----------- |
| **GET** | `/validate?code={code}&redemption_type={redemption_type}&api_key={api_key}`|

Responses
| Status Code | Reason |
| ----------- | ----------- |
| 200 OK | Valid redemption code |
| 400 Bad Request | Missing / Invalid query parameter |
| 401 Unauthorized | Invalid API key |
| 404 Not Found | Invalid redemption code |
| 500 Server Error | Internal server error |
---
**redeem** - redeems instant reward or coupon code
| Verb | Endpoint |
| ----------- | ----------- |
| **GET** | `/redeem?id={id}&redemption_type={redemption_type}&api_key={api_key}`|

## Deployment

Add file `creds.yml` to root of project folder with the following credentials:
```
DB_HOST: XXX
DB_USER: XXX
DB_PASSWORD: XXX
DB_NAME: XXX

```

### Dev

`sls deploy`

### Prod

`sls deploy --stage prod`


