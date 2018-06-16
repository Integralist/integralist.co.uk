---
title: "Authentication with AWS Cognito"
date: 2018-06-15T17:44:31+01:00
categories:
  - "code"
  - "guide"
  - "security"
tags:
  - "authentication"
  - "aws"
  - "cognito"
  - "jwt"
draft: false
---

- [Introduction](#introduction)
- [Requirements](#requirements)
- [Authentication vs Authorization](#authentication-vs-authorization)
- [AWS Cognito](#aws-cognito)
- [User Pools vs Identity Pools](#user-pools-vs-identity-pools)
- [Client-Side vs Server-Side](#client-side-vs-server-side)
- [Social Providers](#social-providers)
- [Hosted UI](#hosted-ui)
- [Migrating Users with AWS Lambda](#migrating-users-with-aws-lambda)
- [Authenticating Cognito Users](#authenticating-cognito-users)
- [JWTs](#jwts)
- [Architecture](#architecture)
- [Transmitting the Tokens](#transmitting-the-tokens)
- [AWS Resources](#aws-resources)
- [Example Cognito App Settings](#example-cognito-app-settings)
- [Example Cognito User Pool “Federation: Identity Providers”](#example-cognito-user-pool-federation-identity-providers)
- [Example Facebook App Configuration](#example-facebook-app-configuration)
- [Example Google App Configuration](#example-google-app-configuration)
- [Example Python Lambda Migration Script](#example-python-lambda-migration-script)
- [Conclusion](#conclusion)

## Introduction

In this post I want to look at how you might implement an authentication system using a third-party provider (specifically AWS Cognito), and although I'll be discussing an imaginary system, the premise is based on a real scenario and set of requirements (so the experiences of figuring out this stuff is pretty much how they happened).

What are the various moving pieces and what options do we have available to us in order to build this system, and how does it integrate with our existing systems?

### Sneak Peek

Here's a quick peek at what we'll be discussing...

<a href="../../images/cognito-high-level.png">
  <img src="../../images/cognito-high-level.png">
</a>

...this will all become clearer as we go on.

## Requirements

Imagine we have a system in place that requires authentication (a Content Managment System is an obvious example). We want access to this system to be restricted to only authenticated users.

> Note: in this example the CMS will be a public system, so users can sign-up/register to acquire access to it.

Now also imagine that this system already has authentication built-in and is backed by an existing user database. We want to migrate our existing users over to the new AWS Cognito authentication system, and to then replace the legacy authentication system.

We could 'proxy' unauthenticated users to Cognito or we could 'redirect' them, either way we'll check if the user is authenticated (we'll explain how that works later) and if not, then we'll give them over to Cognito to handle things from there and to send the authenticated user back to us (i.e. the CMS).

## Authentication vs Authorization

It's important to clarify that in this blog post we're only really discussing authentication, and not _authorization_. They are two different concepts.

- **Authentication** is the process of verification that an individual, entity or website is who it claims to be.

- **Authorization** is the function of specifying access rights to resources, which is different to (and commonly confused with) the process of authentication.

> Note: refer to [this glossary document](https://docs.google.com/document/d/1qs3jEIQvocdVhSxCSPLF1BoLnp91aLnuUIasvl-maYo/edit#) for a list of security terminology.

## AWS Cognito

First, let's discuss what Amazon Cognito is. According to [the official blurb](https://aws.amazon.com/cognito/)...

> Amazon Cognito lets you add user sign-up, sign-in, and access control to your web and mobile apps quickly and easily.

That sounds great. So what exactly does it offer?

## User Pools vs Identity Pools

There are two sides to AWS Cognito:

1. User Pools
2. Identity Pools

User Pools deal with 'authentication', where as Identity Pools deal with 'authorization' (and specifically that means AWS based resources only).

For the purposes of this post, we'll only be focusing in on User Pools as we don't need to worry about authenticated CMS users having authorized access to other AWS resources like S3 or whatever.

> If you're interested in the various Identity Pool concepts, then please refer to [the official documentation](https://docs.aws.amazon.com/cognito/latest/developerguide/authentication-flow.html).

## Client-Side vs Server-Side

So you'll find that AWS has a bias towards implementing Cognito using client-side technology. Specifically using their [Amplify](https://aws.github.io/aws-amplify/) JS SDK (and by that I mean, they push it _hard_ whenever they can -- they advised against us using a server-side language or building our own UI etc).

> Note: the AWS docs are out of date in areas, and so you'll likely see it reference [amazon-cognito-identity-js](https://github.com/amazon-archives/amazon-cognito-identity-js) which is now deprecated, and actually Amplify integrated it with its own abstraction put on top.

AWS, I feel, realised that using Cognito isn't as straight forward as they make out in their promotional blurb and so decided to do something about it and put a lot of time and effort into this new Amplify JS SDK. 

Unfortunately the other SDK's they provide don't have the same level of built-in functionality provided (I guess simply because they have more flexibility on the client-side?) and so if you want to use Python or Go (or whatever else), then the steps to build something on top of Cognito are more manual.

To give some background context to why I'm mentioning this: I started by playing around with Amplify, as I thought it would be quick and simple to get up and running. It wasn't. 

But to be fair, that's because I don't know the modern JavaScript landscape, and unfortunately picking up and learning the various frameworks like Angular or React (+ WebPack and other build systems etc...) isn't trivial.

I moved onto building my own custom UI (just simple stuff) that would interact with Cognito using the Python SDK. This worked fine with User Pools, but when it came time to implement social logins (facebook, google etc) using the SDK it became apparent that this wasn't as straight forward as I imagined (e.g. just call an API endpoint, passing some credentials and let Cognito figure out the rest).

## Social Providers

In order to implement social logins (e.g. facebook or google), you have a few options and each has their own set of considerations:

1. Amplify
2. Server-Side SDK's
3. AWS Hosted UI

We already decided that option 1. was off the table due to us being out of our comfort zone with the client-side tech stack. Authentication is important and you should not be relying on hosting/running code that you don't understand when it comes to something important like that. Either delegate the responsibility or understand the code.

Option 2. was what we tried next, and we spent a fair amount of time building out our own ui, as well as setting up and testing user authentication with a User Pool. But by the time we started to integrate social logins we realised that the SDK's don't offer any API calls for handling social logins 🤦♂️ .

Unfortunately it wasn't until _after_ we had moved to option 3. (AWS' own self hosted ui) that we realised there _was_ a way to do this server-side using a specific [User Pools "Auth API"](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-userpools-server-contract-reference.html). 

The reason we didn't discover that particular auth api request flow immediately was because "User Pools" was something we initially had associated with the traditional username/password type accounts stored in Cognito and not social provider logins (that was our mistake).

As it turns out, you _can_ set-up users so they can authenticate using a social provider, but you need to first configure a 'client application' in AWS. Once you have an application created you can configure various social providers via AWS' federated identity providers.

> Note: if things don't work out using Cognito's hosted ui, then we might well move back to the server-side SDK's so we have more control over the UI aspect.

## Hosted UI

So option 3. is where we're at currently, and it works well so far, but it does have some caveats:

- Very limited controls over the ui (_very_ basic font colors and css).
- Custom domains only work with TLS certificates via [ACM](https://aws.amazon.com/certificate-manager/).
- Resetting password only accepts username, not an email.

> Note: there are other issues I have with the hosted ui, but in a lot of cases it does the job well enough to put up with them.

## Migrating Users with AWS Lambda

As I mentioned earlier, our CMS is backed by a datastore of users and we want to move away from the simple authentication logic implemented within the CMS with AWS Cognito. So how do we do that?

AWS discusses various approaches in their [blog](https://aws.amazon.com/blogs/mobile/migrating-users-to-amazon-cognito-user-pools/), but for us the "one-at-a-time user migration" is the approach that works best. The way it works is like so:

- Existing user attempts sign-in.
- Cognito can't find the user so fails the login.
- Cognito triggers a lambda.
- The lambda is passed the user's credentials.
- Your lambda code authenticates the user with your original datastore.
- You take the returned user details and create a new 'cognito' user.
- Next time user authenticates, cognito will find them in the user pool.

## Authenticating Cognito Users

Once we authenticate a user with Cognito, the user will be redirected to another uri. Typically this would be an endpoint hosted by _your_ system that required the user authentication in the first place (for our purposes, this would be our CMS). Something like `/auth/signin/callback/`.

When that uri is called, it'll be passed a 'code' query param (e.g. `/auth/signin/callback?code=<...>`). Your service now needs to _exchange_ that code for a 'user pool' token. The token you receive back will be a [JWT](https://jwt.io/). 

> Note: if you're unfamiliar with the concept of JWTs, then I highly recommend reading [jwt.io/introduction](https://jwt.io/introduction/).

## JWTs

When you exchange the cognito 'code' for a user pool 'token', you'll actually be returned _three_ tokens:

1. ID token
2. Access token
3. Refresh token

> Note: see documentation for more details on [these three tokens](https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-with-identity-providers.html).

The ID token provides details about the user, and the access token indicates the access allowed to that user's attributes stored within the Cognito User Pool. 

Both the ID token and access token will expire after one hour. To use them after that you'll need the refresh token to refresh the access/id tokens for another hour. The refresh token expires after 30 days.

> Note: where you make the 'refresh' API calls from is up to you. You could do from within the CMS or you could push out that functionality to some kind of API service (which is what we ultimately did).

In our example scenario, the CMS will need the ID token in order to confirm the user has authenticated (we'll come back to how that works later).

## Architecture

<a href="../../images/cognito-high-level.png">
  <img src="../../images/cognito-high-level.png">
</a>

So we took a sneak peek at this architecture earlier, and now that we understand a bit more about the Cognito landscape we can start to break down this diagram:

1. We redirect an unauthenticated user to Cognito.
2. Once the user attempts to sign-in we trigger some additional '[hooks](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-identity-pools-working-with-aws-lambda-triggers.html)'.
3. Cognito redirects the authenticated user to our API service † 
4. Our API service redirects the user back to the CMS (with user tokens).
5. The CMS asks the API service to validate the tokens.

> † this service exchanges the given Cognito auth code for the user's Cognito User Pool tokens.

Once the tokens are validated, the CMS will allow the user to view the relevant page.

## Transmitting the Tokens

Our API service is responsible for handling the callback endpoint that Cognito calls once a user successfully authenticates. Once the API service successfully exchanges the Cognito 'code' for User Pool tokens it will store the tokens into a cookie.

> Note: we like to set [`Secure`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#Secure_and_HttpOnly_cookies) and [`HttpOnly`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#Secure_and_HttpOnly_cookies) on our cookies (you don't have to, but we like to).

We do this so we can transmit the tokens to the CMS.

Specifically we'll set the cookie with a wildcard domain so that we can access the cookie in our CMS service. The reason this works is because although the CMS and API are two separate services, they're both hosted off the same TLD (e.g. `cms.example.com` and `auth-api.example.com`).

## AWS Resources

There are a few resources that need to be set-up:

- IAM user (with access to Cognito APIs)
- Cognito User Pool
- Lambda x2
  - User Migration hook
  - Post authentication hook

> Note: both lambda's need an IAM role with access to CloudWatch Logs and Cognito's APIs

Here is an example IAM role policy for the lambda's to use:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:aws:logs:*:*:*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "cognito-idp:AdminUpdateUserAttributes"
            ],
            "Resource": "arn:aws:cognito-idp:us-east-1:{aws_account_id}:userpool/{user_pool_id}"
        }
    ]
}
```

> Note: don't forget to update `aws_account_id` and `user_pool_id` in the code snippet.

## Example Cognito App Settings

This isn't meant to be an exhaustive example, but it gives you an idea of some of the configuration you'll need.

- **Callback URL(s)**:  
  `https://auth-api.example.com/auth/signin/callback`

- **Allowed OAuth Flows**:  
  Authorization code grant  
  Implicit grant

- **Allowed OAuth Scopes**:  
  email  
  openid  
  profile

## Example Cognito User Pool "Federation: Identity Providers"

For each provider there is a "Authorize Scope" section.

- **Facebook**:  
  public_profile,email

- **Google**:  
  profile email openid

## Example Facebook App Configuration

https://developers.facebook.com/apps

- **App Domains**:  
  `https://your-organisation.auth.us-east-1.amazoncognito.com`

- **Privacy Policy URL**:  
  `https://www.example.com/about/privacy`

- **Site URL**:  
  `https://your-organisation.auth.us-east-1.amazoncognito.com/oauth2/idpresponse`

### Product Added: "Facebook Login"

- **Client OAuth Login**:  
  Yes

- **Web OAuth Login**:  
  Yes

- **Enforce HTTPS**:  
  Yes

- **Valid OAuth Redirect URIs**:  
  `https://your-organisation.auth.us-east-1.amazoncognito.com/oauth2/idpresponse`  
  `https://auth-api.example.com/auth/signin/callback`

## Example Google App Configuration

https://console.developers.google.com/

- **Enabled API(s)**:  
  Google+ API

- **Credentials Type**:  
  OAuth client ID

- **Application Type**:  
  Web application

- **Authorized JavaScript origins**:  
  `https://your-organisation.auth.us-east-1.amazoncognito.com`

- **Authorized redirect URIs**:  
  `https://your-organisation.auth.us-east-1.amazoncognito.com/oauth2/idpresponse`  
  `https://auth-api.example.com/auth/signin/callback`

## Example Python Lambda Migration Script

The code is effectively made up, but gives an idea of what you might execute within this 'user migration' lambda hook.

But one there is one aspect that's based off a real scenario, which is taking the returned authentication details from the legacy system `legacy_id` and storing that as a 'custom attribute' within Cognito's user pool for the user we were about to migrate.

We do this so that the next time this user authenticates, our existing CMS code can be modified to lookup the user's data from Cognito. You might wonder why that's even necessary? Well for us we have mobile applications relying on our legacy authentication system and they won't be able to migrate to Cognito for a few months.

Meaning: we have two authentication systems that need to stay functioning alongside each other for the foreseeable future. Your mileage may vary.

```
from botocore.vendored import requests


def authn_legacy(event):
    data = {'username': event['userName'],
            'password': event['request']['password']}

    uri = f'https://cms.example.com/api/login'

    return requests.post(uri, data=data)


def lambda_handler(event, context):
    """
    There are two states:

    1. authentication (but before user created in cognito)
    2. forgotten password (we don't cover that in this example code)

    For authentication we log the user into the legacy authentication system 
    and then extract their details, then modify the user account that is 
    about to be created in cognito's user pool.
    """

    if event['triggerSource'] == "UserMigration_Authentication":
        login_response = authn_legacy(event)

        if login_response.status_code == 200:
            d = login_response.json()

            user_attr = {'username': d['username'],
                         'name': d['display_name'],
                         'email': d['email'],
                         'email_verified': 'true',
                         'custom:legacy_id': d['userid']}

            event['response']['userAttributes'] = user_attr
            event['response']['finalUserStatus'] = "CONFIRMED"
            event['response']['messageAction'] = "SUPPRESS"

            return event
        else:
            return None
    elif event['triggerSource'] == "UserMigration_ForgotPassword":
        # TODO: an exercise for the reader
        return None
    else:
        return None
```

Something else of interest in this code is that when the lambda finishes executing, as long as the response isn't `None`, Cognito will create a new user. So we explicitly return the original `event` that was passed to the lambda.

The event we return has a `response` setting where we can make modifications to the user account about to be created within Cognito.

> Note: you'll find that a social login, doesn't cause the 'user migration' hook to fire, and so we found we needed to do some extra work for social users and that meant using another lambda trigger ('post authentication').

## Conclusion

I personally found the documentation around Cognito (and the various tools) to be both overwhelming and underwhelming. Not to mention confusing in places, as well as just downright frustrating at times. 

Hopefully you found this short break down of AWS Cognito useful. There's so much more still to dive into, but this should give you at least a decent starting point.
