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
- [Architecture](#architecture)
- [Authenticating Cognito Users](#authenticating-cognito-users)
- [JWTs](#jwts)
- [Transmitting the Tokens](#transmitting-the-tokens)
- [Validating Authenticated Users](#validating-authenticated-users)
- [AWS Resources](#aws-resources)
- [Example Cognito App Settings](#example-cognito-app-settings)
- [Example Cognito User Pool “Federation: Identity Providers”](#example-cognito-user-pool-federation-identity-providers)
- [Example Facebook App Configuration](#example-facebook-app-configuration)
- [Example Google App Configuration](#example-google-app-configuration)
- [Example Python Lambda Migration Script](#example-python-lambda-migration-script)
- [Conclusion](#conclusion)

## Introduction

In this post I want to talk you through how we started to investigate how we might migrate a legacy authentication system over to a new provider (specifically AWS Cognito). The majority of the architecture and code discussed here is made-up, but there is still enough of a similarity to the actual system to hopefully be useful to others in a similar position.

My goal is to share with you the learning experiences I had and to help others understand the various moving pieces (and the various options) available to us in order to build this type of system, and how it might integrate with our existing legacy systems.

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

### SRP

One caveat of doing authentication server-side was that (depending on the APIs you used) it could require some pretty intense code for generating hashes, hexes and large numbers in order to be passed along with the API calls you were making.

An example of this is the [`InitiateAuth`](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_InitiateAuth.html) API call with the `USER_SRP_AUTH` auth flow, which requires you to generate an `SRP_A`. I won't go into the details (read the docs for more info) but take a look at a popular Python Cognito abstraction library called [Warrant](https://github.com/capless/warrant/blob/master/warrant/aws_srp.py) to get an idea of the amount of code required to be written for what otherwise seems like a simple API call. 

It wasn't until later, we discovered that we could (in the case of `InitiateAuth` at least) have avoided writing all the SRP generation code and instead used an _admin_ API call instead called [`AdminInitiateAuth`](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminInitiateAuth.html) (you skip SRP in favour of implicitly trusting the API calls -- e.g. they're coming from a trusted administrator/service). 

> Note: it seems knowing _when_ you should use `InitiateAuth` over `AdminInitiateAuth` is unclear to me (in the case of server-side interactions I'm not sure why you wouldn't always just use the admin API call).

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

The hosted ui option gives us all the interactions (along with a fully functioning ui) for free, this includes: sign-in, sign-up, forgotten username, forgotten password, social logins.

The uri for the hosted ui looks something like:

`https://your-organisation.auth.us-east-1.amazoncognito.com/login?response_type=code&client_id=<...>&redirect_uri=<...>&state=<...>`

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

### Lambda Flow

It's worth noting the second half of the diagram (the section _after_ the lambda is triggered). What we have here are two separate lambda's, and which one is triggered depends on the scenario. 

If the user has tried to authenticate using credentials that don't exist in the Cognito User Pool, then the "User Migration" lambda is triggered and once we authenticate the user within our legacy system (that is the call over to the "WebApp" in the diagram), then Cognito will create the user within its User Pool.

Annoyingly we found that a social login (e.g. someone logs in using their Facebook or Google account instead of a traditional username/password), doesn't cause the 'User Migration' hook to be triggered. But we needed to process some logic when a social user attempts to sign-in and so the 'Post Authentication' hook allowed us to do that.

> Note: we'll explain this social login issue in more detail in the next section.

## Legacy System

The problem with the social user sign-in not triggering the 'User Migration' lambda is worth elaborating upon because those users might not exist within our legacy system, and to be sure that we can migrate our authentication layer over to Cognito means not breaking existing legacy code that attempts to look up user data from within our legacy user datastore.

To prevent our code from breaking we need to ensure there is a row in the legacy datastore for that new user, so we create the a legacy user account while the lambda is executing. 

When we create the user account in the legacy datastore we'll get an id back, and so within the lambda we'll make sure to associate that with the newly created Cognito social user account (which is created within our User Pool under either a Facebook or Google user 'group') via a custom attribute: `legacy_uid`. 

Doing this allows us to correlate a Cognito user with a legacy user and we also already do this for normal username/password authenticated users.

We also do the _reverse_: storing the Cognito user id into our legacy system (i.e. users who don't exist within our legacy system get a legacy account created for them even though they exist primarily as a new Cognito user). 

What this means is that when we create the legacy user account, we make sure to add into a new column the Cognito user id. 

Before we transition to the new Cognito system, we'll need to ensure that any areas of our legacy UI that allows a user to update their details (stuff that's related to authentication like their 'email') is also synced with their details stored within the Cognito User Pool.

## Authenticating Cognito Users

Once we authenticate a user with Cognito, the user will be redirected to another uri. Typically this would be an endpoint hosted by _your_ system that required the user authentication in the first place (for our purposes, this would be our CMS). Something like `/auth/signin/callback/`.

When that uri is called, it'll be passed a 'code' query param (e.g. `/auth/signin/callback?code=<...>`). Your service now needs to _exchange_ that code for a 'user pool' token. The token you receive back will be a [JWT](https://jwt.io/). 

> Note: if you're unfamiliar with the concept of JWTs, then I highly recommend reading [jwt.io/introduction](https://jwt.io/introduction/).

Here's an example of how the 'code for token' exchange works:

```
# set the following to relevant values
client_id = <auth_api_app_client_id>
secret_key = <auth_api_app_secret_key>

# this should be the tld and path of the
# callback endpoint, which as described
# earlier was something like:
# https://auth-api.example.com/auth/signin/callback
redirect_host = <redirect_host>
redirect_path = <redirect_path>

uri = 'https://your-organisation.auth.us-east-1.amazoncognito.com/oauth2/token'

code = self.get_query_argument('code', '')
state = self.get_query_argument('state', '')

m = re.search('redirect=(.*)', state)
redirect = ''
if m:
    redirect = m.group(1)

data = {'code': code,
        'grant_type': 'authorization_code',
        'client_id': client_id,
        'redirect_uri': f'{redirect_host}{redirect_path}'}

r = requests.post(uri, data=data, auth=(client_id, secret_key))

tokens = r.json()
id_token = tokens.get('id_token')
access_token = tokens.get('access_token')
refresh_token = tokens.get('refresh_token')

if not id_token or not access_token or not refresh_token:
    logging.error('login failed: missing token(s)')
else:
    cookie_name = 'usercookie'
    cookie_value = f'id_token={id_token};access_token={access_token};refresh_token={refresh_token}'
    cookie_args = {'secure': True, 'httponly': True}

    self.set_cookie(cookie_name,
                    cookie_value,
                    domain=<cookie_domain>,
                    expires_days=30,
                    **cookie_args)

if redirect:
    self.redirect(redirect, status=302)
else:
    self.redirect(redirect_host, status=302)
```

So there are a few `<...>` items in that code where you would change the value to something relevant to your use case, but ultimately this acquires the `code` query param value and sends it to the Cognito hosted ui endpoint:

`https://your-organisation.auth.us-east-1.amazoncognito.com/oauth2/token` 

...and as long as you provide the correct client application credentials, you'll get the user pool tokens in return.

Now there are some confusing aspects to this code we'll cover briefly:

1. state query param
2. parsing a 'redirect'
3. cookie size

### State for CSRF

Remember the hosted ui endpoint was something like:

`https://your-organisation.auth.us-east-1.amazoncognito.com/login?response_type=code&client_id=<...>&redirect_uri=<...>&state=<...>`

Notice the `state` param. That should be assigned a random value, which acts like a CSRF token, and so is the only given query param that persists the redirect to your callback endpoint.

### Overloading the State Param

The problem we had in our real life scenario was that we _didn't_ redirect back to our main domain (e.g. cms.example.com) and we didn't have that service handle the exchange of the 'code for user pool tokens' before redirecting to the endpoint that originally required authentication. 

Instead we redirected from the hosted ui to a `auth-api.example.com` service that was responsible for handling the code exchange, before redirecting back to the `cms.example.com` endpoint.

The problem with our setup is that we in effect have _two_ separate redirects:

1. cognito hosted ui -> auth-api.example.com
2. auth-api.example.com -> cms.example.com

But cognito's hosted ui only provides us a single `redirect_uri` endpoint, and we don't want our `auth-api.example.com` service to have to handle any state other than what's provided to its endpoints (in this case the `code` param and its value which we would exchange for a user token).

We couldn't hardcode into the `auth-api.example.com` service the (for example) `cms.example.com/new-post` endpoint (which you would see once you were authenticated) because in essence we intended on implementing this same hosted ui in front of multiple microservices each with their own endpoints that require authentication.

So we ended up in a scenario that a few other companies have stumbled into, where we were asking:

**How do we provide _two_ redirects instead of just one?**

Well, we do what those other companies have apparently done, and overloaded the `state` param to have a value of something like: `state=1234_redirect=https://cms.example.com/new-post`. 

Once we do that, the above code for `/auth/signin/callback` is able to extract the redirect value and make sure that once we exchange our code for a token, that we're able to then redirect to the correct final destination.

> Note: in reality we have a white list of domains that we validate the param value against so that this behaviour can't be manipulated.

### Problems with Cookie Size

Another problem we stumbled across was setting a single cookie containing _all_ the user pool tokens (see the next section), and so we would reach a 'request header size limit exceeded' error. 

What we decided to do was set multiple cookies to account for the size issue.

## JWTs

When you exchange the cognito 'code' for a user pool 'token', you'll actually be returned _three_ tokens:

1. ID token
2. Access token
3. Refresh token

> Note: see documentation for more details on [these three tokens](https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-with-identity-providers.html).

The ID token provides details about the user, and the access token indicates the access allowed to that user's attributes stored within the Cognito User Pool. 

Both the ID token and access token will expire after one hour. To use them after that you'll need the refresh token to refresh the access/id tokens for another hour. The refresh token expires after 30 days.

> Note: where you make the 'refresh' API calls from is up to you. You could do it from within the CMS or you could push out that functionality to some kind of API service (which is what we ultimately did).

In our example scenario, the CMS will need the ID token in order to confirm the user has authenticated (we'll come back to how that works later).

## Transmitting the Tokens

Our API service is responsible for handling the callback endpoint that Cognito calls once a user successfully authenticates. Once the API service successfully exchanges the Cognito 'code' for User Pool tokens it will store the tokens into a cookie.

> Note: we like to set [`Secure`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#Secure_and_HttpOnly_cookies) and [`HttpOnly`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#Secure_and_HttpOnly_cookies) on our cookies (you don't have to, but we like to).

We do this so we can transmit the tokens to the CMS.

Specifically we'll set the cookie with a wildcard domain so that we can access the cookie in our CMS service. The reason this works is because although the CMS and API are two separate services, they're both hosted off the same TLD (e.g. `cms.example.com` and `auth-api.example.com`).

## Validating Authenticated Users

The `cms.example.com` endpoint(s) that require authentication would need to look for a cookie that contains the JWT user pool tokens. Once found, and the tokens extracted, it would pass them to our `auth-api.example.com` service to be validated using AWS's [JWK](https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-with-identity-providers.html#amazon-cognito-identity-user-pools-using-id-and-access-tokens-in-web-api).

This is cryptography in action, the JWK is a set of keys that are the public equivalent of the private keys used by AWS to digitally sign the tokens. We acquire these via a standard format endpoint:

`https://cognito-idp.{region}.amazonaws.com/{userPoolId}/.well-known/jwks.json.`

The code for this would look something like:

```
def check_token(access_token: str):
    if not access_token:
        raise Exception('Access Token Required to Check Token')

    now = datetime.datetime.now()
    dec_access_token = jwt.get_unverified_claims(access_token)

    if now > datetime.datetime.fromtimestamp(dec_access_token['exp']):
        expired = True
    else:
        expired = False

    return expired


def secret_hash(username, client_id, client_secret):
    message = bytearray(username + client_id, 'utf-8')
    hmac_obj = hmac.new(bytearray(client_secret, 'utf-8'), message, hashlib.sha256)
    return base64.standard_b64encode(hmac_obj.digest()).decode('utf-8')

id_token = self.get_arguments('id_token')
access_token = self.get_arguments('access_token')
refresh_token = self.get_arguments('refresh_token')

if not id_token or not access_token or not refresh_token:
    logging.error('if not id_token or not access_token or not refresh_token')
    self.write({'verified': False})
    return

try:
    payload = jwt.decode(id_token[0].encode('utf-8'), verify=False)
except Exception:
    logging.error('jwt decode error')
    self.write({'verified': False})
    return

user = payload['cognito:username']

try:
    if check_token(access_token=access_token[0]):
        params = {'REFRESH_TOKEN': refresh_token[0],
                  'SECRET_HASH': secret_hash(user, client_id, client_secret)}

        client.initiate_auth(ClientId=client_id,
                             AuthFlow='REFRESH_TOKEN',
                             AuthParameters=params)
except Exception as err:
    logging.error(err)
    self.write({'verified': False})
    return

self.write({'verified': True, 'user': payload})
```

> Note: we would also update the cookie that holds the old JWT user pool tokens, in the case of the tokens being updated after verification initially failed due to expiry of the provided tokens (as long as the refresh token hadn't also expired).

## AWS Resources

OK, let's take a moment to consider some of the AWS resources that are needed for this set-up:

- IAM user (with access to Cognito APIs)
- Cognito User Pool
- Lambda x2
  - User Migration hook
  - Post authentication hook

> Note: both lambda's need an IAM role with access to CloudWatch Logs and Cognito's APIs

### IAM User

The IAM _user_ is necessary as we have to provide some credentials to the [boto client](https://boto3.readthedocs.io/en/latest/reference/services/cognito-idp.html) ([boto](https://boto3.readthedocs.io/en/latest/reference/core/boto3.html) is the Python SDK) in order for it to make certain API calls.

Here's an example of the code to instantiate a client:

```
client = boto3.client('cognito-idp', **{'aws_access_key_id': access_key,
                                        'aws_secret_access_key': secret_key,
                                        'region_name': region})
```

Notice the service name is `cognito-idp` and not `cognito-identity`. I mention this as the docs specify two different services "CognitoIdentity" and "CognitoIdentityProvider", which when we were first learning about Cognito we presumed the latter "CognitoIdentityProvider" was something associated with Cognito Identity Pools. 

As we were only interested in the User Pool functionality, we found it strange that the small number of examples we found online all referenced `cognito-idp`. 

So we struggled for a bit to understand the difference, and although we used the "CognitoIdentityProvider" service (i.e. `cognito-idp`), we were confused for a long time as to why that was the case.

Turns out that Cognito's "User Pool" is itself fundamentally a _identity provider_, and because of that you can configure a "Identity Pool" to have a "User Pool" associated within it (along with more common external identity providers such as Facebook and Google).

### Lambda IAM Role

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

It simply sets up CloudWatch logs access, and allows us (as an 'admin') to update user attributes within our User Pool.

> Note: if you're 'copying and pasting', don't forget to update `aws_account_id` and `user_pool_id` in the code snippet.

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

The following code is made up, but gives an idea of what the actual implementation might resemble for the 'user migration' lambda hook.

But there is one aspect that's based off a real scenario, which is taking the returned authentication details from the legacy system `legacy_id` and storing that as a 'custom attribute' within Cognito's user pool for the user we were about to migrate.

We do this so that the next time this user authenticates, our existing CMS code can be modified to lookup the user's data from Cognito. You might wonder why that's even necessary? Well for us we have mobile applications relying on our legacy authentication system and they won't be able to migrate to Cognito for a few months.

Meaning: we have two authentication systems that need to stay functioning alongside each other for the foreseeable future. Your mileage may vary.

> Note: our legacy system would first check for the cookie that potentially contains our authenticated user's tokens, and if that wasn't found it would look for a legacy session cookie instead (if _that_ didn't exist, then it would redirect to the Cognito hosted ui so the unauthenticated user could authenticate appropriately).

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

## Example Python Lambda Social Script

Remember the purpose of this script is to create a new social user within our legacy datastore and to also ensure that the `legacy_uid` and Cognito user id are stored in both authentication datastores so we're able to cross reference users while we migrate more legacy systems over to Cognito.

```
import os
import random
import re
import string
import boto3

from botocore.vendored import requests


def social_account_lookup(provider, uid):
    """See if the webapp has a user account with this social provider."""

    url = f'https://cms.example.com/api/cognito_helper?provider={provider}&user_id={uid}'
    response = requests.get(url)
    return response.json()


def generate_random_output():
    return ''.join(random.sample(string.ascii_lowercase, 10))


def create_legacy_user(provider_name, uid, display_name, email, username, rand_pass):
    """Create new legacy user for the authenticated social user."""

    params = f'?{provider_name}={uid}&name={display_name}&username={username}&email={email}&pw={rand_pass}'
    url = f'https://cms.example.com/api/create_legacy_account{params}'

    response = requests.get(url)
    response_data = response.json()

    if response_data['status'] == 'failed':
        return None

    return response_data['userid']


def update_legacy_user(legacy_uid, cognito_id):
    """Update webapp account so it has a cognito id associated with it."""

    url = f'{https://cms.example.com/api/cognito_helper?cognito_id={cognito_id}&user_id={legacy_uid}'

    response = requests.get(url)
    response_data = response.json()

    if response_data['success'] == 0:
        return None


def no_legacy_account(response):
    return response['success'] == 0


def modify_cognito_user(username, legacy_uid):
    attributes = [{'Name': 'custom:legacy_uid', 'Value': legacy_uid}]

    # credentials are dynamically pulled from IAM role
    client = boto3.client('cognito-idp', **{'region_name': 'us-east-1'})

    response = client.admin_update_user_attributes(**{'UserPoolId': '...',
                                                      'UserAttributes': attributes,
                                                      'Username': username})


def post_authentication(event):
    return event['triggerSource'] == 'PostAuthentication_Authentication'


def external_provider(event):
    return event['request']['userAttributes']['cognito:user_status'] == 'EXTERNAL_PROVIDER'


def migrated(event):
    """We know an account has been migrated if there is a legacy_uid in cognito."""
    return event['request']['userAttributes'].get('custom:legacy_uid')


def lambda_handler(event, context):
    if post_authentication(event) and external_provider(event) and not migrated(event):
        cognito_id = event['request']['userAttributes']['sub']
        identities = event['request']['userAttributes']['identities']
        provider = re.search('providerName":"(.+?)"', identities)
        uid = re.search('userId":"(.+?)"', identities)

        if provider and uid:
            provider = provider.group(1)
            uid = uid.group(1)

            response = social_account_lookup(provider, uid)

            if no_legacy_account(response):
                display_name = event['request']['userAttributes']['name']
                email = event['request']['userAttributes']['email']
                username = event['userName']
                rand_pass = generate_random_output()

                provider_name = f'{provider.lower()}_uid'
                legacy_uid = create_legacy_user(provider_name, uid, display_name, email, username, rand_pass)

                if not legacy_uid:
                    return None
            else:
                username = response['user']['username']
                legacy_uid = response['user']['id']

            modify_cognito_user(username, legacy_uid)
            update_webapp_user(legacy_uid, cognito_id)
    else:
        user = event['request']['userAttributes']['name']
        legacy_uid = event['request']['userAttributes']['custom:legacy_uid']
        print(f"this user {user} already has a legacy_uid {legacy_uid} in their cognito account")

    return event
```

## Conclusion

There's so much more to the story, but I think this post is long enough as it is and I don't want to keep you any longer. If you have any questions, then please reach out to me on twitter.

I personally found the documentation around Cognito (and the various tools) to be both overwhelming and underwhelming. Not to mention confusing in places, as well as just downright frustrating at times. 

Hopefully you found this short break down of AWS Cognito useful. There's so much more still to dive into, but this should give you at least a decent starting point.
