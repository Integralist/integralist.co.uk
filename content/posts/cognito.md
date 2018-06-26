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
- [What is Cognito?](#what-is-cognito)
- [Authentication vs Authorization](#authentication-vs-authorization)
- [User Pools vs Identity Pools](#user-pools-vs-identity-pools)
- [Implementation Options](#implementation-options)
    - [Client SDK](#client-sdk)
    - [Server SDK](#server-sdk)
    - [AWS Hosted UI](#aws-hosted-ui)
- [Logic Processing with AWS Lambda](#logic-processing-with-aws-lambda)
- [Beware the Lambdas](#beware-the-lambdas)
- [Social Logins](#social-logins)
    - [Overloading the State Parameter](#overloading-the-state-parameter)
- [JWTs](#jwt)
- [API Limits](#api-limits)
- [Which is the right solution?](#which-is-the-right-solution)
- [User Pool Configuration](#user-pool-configuration)
- [IAM User](#iam-user)
- [Lambda IAM Role](#lambda-iam-role)
- [Example Cognito App Settings](#example-cognito-app-settings)
- [Example Cognito User Pool “Federation: Identity Providers”](#example-cognito-user-pool-federation-identity-providers)
- [Example Facebook App Configuration](#example-facebook-app-configuration)
- [Example Google App Configuration](#example-google-app-configuration)
- [Conclusion](#conclusion)

## Introduction

In this post I would like to introduce you to the [AWS Cognito](https://aws.amazon.com/cognito/) service, and to explain its various moving pieces and how they fit together.

I personally found Cognito hard to get up and running with (for various reasons I'll explain as we go), and to make things worse there didn't seem to be that many reference points outside of the official documentation to help me. Hence this blog post now exists for those weary travellers looking for answers.

Let's start at the beginning...

## What is Cognito?

According to [the official blurb](https://aws.amazon.com/cognito/)...

> Amazon Cognito lets you add user sign-up, sign-in, and access control to your web and mobile apps quickly and easily.

In essence, Cognito provides features that let you _authenticate_ access to your services, while also providing features to let you _authorize_ access to your AWS resources.

## Authentication vs Authorization

It's important to clarify that in this blog post we're only really discussing authentication, and not _authorization_. They are two different concepts.

- **Authentication** is the process of verification that an individual, entity or website is who it claims to be.

- **Authorization** is the function of specifying access rights to resources, which is different to (and commonly confused with) the process of authentication.

> Note: if you're new to these types of security concepts, then take a look at [this glossary document](https://docs.google.com/document/d/1qs3jEIQvocdVhSxCSPLF1BoLnp91aLnuUIasvl-maYo/edit#) for various terminology.

## User Pools vs Identity Pools

In order for you to be able to _authenticate_ and _authorize_ access, Cognito provides two separate services:

1. User Pools
2. Identity Pools

User Pools deal with 'authentication', where as Identity Pools deal with 'authorization' (and specifically that means AWS based resources only).

For the purposes of this post I'll only be focusing in on User Pools, as I've not yet had to worry about authorizing access for AWS resources to an authenticated user (which is where Identity Pools would come into play).

> If you're interested in the various Identity Pool concepts, then please refer to [the official documentation](https://docs.aws.amazon.com/cognito/latest/developerguide/authentication-flow.html).

## Implementation Options

There are fundamentally three options available for implementing User Pools:

1. Client SDK
2. Server SDK
3. AWS Hosted UI

### Client SDK

The client SDK has a bit of a jagged history, which makes reading the AWS docs a bit confusing at times (or indeed when Googling for help), as you may notice references to '[Amazon Cognito Identity SDK for JavaScript](https://github.com/amazon-archives/amazon-cognito-identity-js)' which is now a deprecated library.

What you'll want to use instead is their new '[Amplify](https://aws.github.io/aws-amplify/)' SDK, which you'll also find AWS has a strong bias towards (or at least their 'solution architects' push it _really_ hard).

> Note: I get the feeling AWS put a lot more time into Amplify and having it be able to abstract away a lot of the Cognito complexity, that they're keen for consumers to utilise it.

Based on this I decided I would trust their opinion and just try and spin up something that works using Amplify, which unfortunately took a long time and ultimately I ended up dropping the work in favour of a server-side solution.

I don't keep up with the constant changes to the JavaScript landscape, and so I'm not familiar with React (or Angular) which were the two examples the AWS docs (and most example repos) used the majority of the time. So using Amplify required me to first do some reading up on React, Babel, WebPack and a whole host of other tools. It was painful.

In the end we just had too much trouble trying to deal with Node and the various build systems that we decided to drop the work we had done and pivot to a new solution (see next section).

### Server SDK

The server-side solution we chose was to use the [Python SDK](https://boto3.readthedocs.io/en/latest/).

This ended up being a bit of a double edged sword. We were happier with the move to Python, but we really struggled with both the AWS documentation and also the boto3 library documentation that the Python SDK is built upon.

In order to get up and running we initially opted to use a 3rd party abstraction library called '[Warrant](https://github.com/capless/warrant)', which also incidentally helped us to understand the AWS documentation because we were able to reverse-engineer the Warrant code to better understand the boto3 API calls that needed to be made.

> Note: I think that says a lot about AWS documentation. If people need to read through how an abstraction library is using your API, then your documentation must be pretty bad. I would be the first to suggest maybe I'm just too dumb to understand Cognito, but a _lot_ of people across the internet were having the same problems.

Ultimately, Warrant didn't provide all the functionality we needed and so we eventually refactored out Warrant and were back to using the underlying boto3 Python SDK.

It's worth me taking a moment to also explain that some APIs require you to define a specific type of 'authentication flow' which is a security feature, and as far as I understand it, is supposed to help you to more safely access data provided by these APIs.

What I didn't know originally, and was one of the reasons we decided to use a library such as Warrant, was that the code involved with some of these auth flows can be quite complex (I still now struggle to follow exactly what the code does within Warrant when it uses one of these 'flows').

Just to give you an example of the type of code AWS Cognito would expect you to write, take a look at the [`InitiateAuth`](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_InitiateAuth.html) API call with the `USER_SRP_AUTH` auth flow. First of all I don't think it's very clear what is expected to be provided in that documentation alone, but also, take a look at [Warrant's implementation](https://github.com/capless/warrant/blob/master/warrant/aws_srp.py) and specifically how to generate an `SRP_A`, which also doesn't appear to be explained anywhere (no where obvious at least).

> Note: it wasn't until much later, we discovered that we could (in the case of `InitiateAuth` at least) have avoided writing all the SRP generation code and instead used the _admin_ version of that API, called [`AdminInitiateAuth`](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminInitiateAuth.html) which allows you to skip SRP in favour of implicitly trusting the caller.

### AWS Hosted UI

AWS Cognito offers a 'hosted ui', where by you redirect a user to an endpoint such as:

`https://{...}.auth.us-east-1.amazoncognito.com/login?response_type={...}&client_id={...}&redirect_uri={...}&state={...}`

> Note: a custom domain can also be configured, but it requires you use [AWS Certificate Manager](https://aws.amazon.com/certificate-manager/) for the TLS cert.

The hosted ui option gives you all the interactions in a fully functioning interface, which includes: sign-in, sign-up, forgotten username, forgotten password, social logins.

But there are some caveats:

- Very limited controls over the ui (_very_ basic font colors and css).
- Custom domains only work with TLS certificates via [ACM](https://aws.amazon.com/certificate-manager/).
- State parameter overloading
- Can't access new signup passwords †

> † this was necessary for my use case as I needed to co-support a legacy system that wasn't ready to migrate over to Cognito

There are other issues still that I have with the hosted ui, but in a lot of cases it does the job well enough to put up with them.

The 'state' parameter overloading is an interesting issue and I'll come back to that later on when I discuss a little bit about sign-ins with social providers.

## Logic Processing with AWS Lambda

With the hosted ui option you'll likely also need to utilise [AWS Lambda](https://aws.amazon.com/lambda/) in order to do some logic processing. The following diagram demonstrates how we were initially using the hosted ui:

<a href="../../images/cognito-high-level.png">
  <img src="../../images/cognito-high-level.png">
</a>

1. We redirect an unauthenticated user to Cognito.
2. Once the user attempts to sign-in we trigger some additional '[hooks](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-identity-pools-working-with-aws-lambda-triggers.html)'.
3. Cognito redirects the authenticated user to our API service † 
4. Our API service redirects the user back to the CMS (with user tokens).
5. The CMS asks the API service to validate the tokens.

> † this service exchanges the given Cognito auth code for the user's Cognito User Pool tokens.

Once the tokens are validated, the CMS will allow the user to view the relevant page.

## Beware the Lambdas

It's worth noting the second half of the above diagram (the section _after_ the lambda is triggered). What we have there are two separate lambda's, and which one is triggered depends on the scenario. 

If the user has tried to authenticate using a username/password set of credentials, and those don't match an existing user within the Cognito User Pool, then the "User Migration" lambda is triggered. In that lambda we attempt to authenticate the user within our legacy system (that is the call over to the "WebApp" in the diagram). 

If the authentication with the legacy system is successful, then we'll modify the user's User Pool record (which hasn't actually been created yet) to include auth related details we've pulled from their legacy account. We then return the 'event' object provided to Lambda, which let's Cognito know it can now create the user within its User Pool (not returning the lambda event object indicates an error occurred and the whole request flow fails).

> Note: with the 'user migration' for users from our legacy system over to Cognito, before we return the event in the lambda, we make sure to mark the new Cognito user as 'verified/confirmed' -- that way they don't need to enter a verification code that gets emailed or sent via SMS (that's because the user would've already verified themselves originally in our legacy system).

The reason I say "beware the lambdas" is because yes, code errors can cause it to bomb out, but more importantly they don't always fire when you think they will (this is a user error thing, not an AWS bug). 

To clarify, let me explain what we saw when testing the migration path of a legacy user account to Cognito, when the user was signing into Cognito using their social provider details.

We had hoped the 'User Migration' lambda hook would have been triggered by both a Cognito User Pool account login and also a Social Provider account login, but it doesn't.

> Note: when a user signs-in with a social account, they have an account created within the Cognito User Pool, but they are also added to a specific group (such as a Facebook group or a Google group).

We _eventually_ discovered that the 'Post Confirmation' hook would fire at the right interval for us to do the processing we needed for users signing in with a Social Account. But that wasn't immediately obvious.

Before settling on the 'Post Confirmation' hook, we originally started using 'Post Authentication' for handling first time social logins (the hook sounded reasonable enough), but when we were testing this hook we already had the social account stored in our User Pool (this was from earlier testing, before we decided to do some 'post-login' processing). 

The reason I mention this is because a week later we decided to clear out our User Pool and start testing our various scenarios again from scratch, and we noticed the 'post authentication' hook was no longer firing 🤔

Turns out social accounts only trigger 'post migration' hooks when they already exist in the User Pool. In order to do the 'first time login' modification we were looking for, we needed the 'post confirmation' hook. 

Using this hook wasn't _obvious_ to us because 'post confirmation' makes it sounds like an event that happens once a username/password user has entered their 'verification code' for the first time (and thus become marked as 'confirmed' within the User Pool). Well, turns out social provider logins are automatically considered _confirmed_ once they authenticate for the first time (hence why that event would trigger when we needed it to).

## Social Logins

One thing that might not be clear when opting for a server-side solution is how to handle social logins (e.g. users signing in/up using facebook or google).

It might sound a bit strange, but in order to implement social logins you'll need to make a call to the hosted ui endpoint (mentioned earlier):

`https://{...}.auth.us-east-1.amazoncognito.com/login?response_type={...}&client_id={...}&redirect_uri={...}&state={...}`

The specific endpoint you call will be based upon those supported in Cognito's [User Pools Auth API](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-userpools-server-contract-reference.html).

For example, to attempt to sign-in a user with facebook you would provide a button that links to:

`https://{...}.auth.us-east-1.amazoncognito.com/oauth2/authorize?response_type={...}&client_id={...}&redirect_uri={...}&state={...}&identity_provider={...}`

The value we use for the `response_type` parameter is `code`. What this does, once the user has authenticated with their social provider (defined by the `identity_provider` param), is redirect the user back to your service (specified via the `redirect_uri` param) and then your service is responsible for exchanging the code for the user's User Pool Tokens (see the following section on [JWTs](#jwts)).

The values you can assign to `identity_provider` are:

- `Facebook`
- `Google`
- `LoginWithAmazon`
    
> Note: if you were planning on handling authentication at a very low level (instead of an SDK), then for a User Pool login you would provide the value `COGNITO`.

### Overloading the State Parameter

The `state` param is used for [CSRF](https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)) protection, and is the only parameter that is persisted when the user is redirected to `redirect_uri`.

A common problem for people using Cognito is that they need more than one redirect. In my case (see the earlier 'hosted ui' architecture diagram) I need to redirect the signed-in user to an API service so we can handle the exchanging of the AWS 'code' for the Cognito User Pool tokens before needing to then redirect the user back to our actual origin service.

The only way we can do this is to overload the `state` param so it has a value like:

`state=123_redirect=https://www.example.com`

The value `123` is the nonce (for CSRF) and the `_` gives us a way to split the query param server-side to extract the secondary redirect endpoint.

> Note: it's recommended you do validation on that input (e.g. a whitelist of accepted URIs) so hackers can't manipulate the endpoint a user is sent to once they've authenticated.

## JWTs

When you exchange the cognito 'code' for a user pool 'token', you'll actually be returned _three_ tokens:

1. ID token
2. Access token
3. Refresh token

> Note: see documentation for more details on [these three tokens](https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-with-identity-providers.html).

The ID token provides details about the user, and the access token indicates the access allowed to that user's attributes stored within the Cognito User Pool. 

Both the ID token and access token will expire after one hour. To use them after that you'll need the refresh token to refresh the access/id tokens for another hour. The refresh token expires after 30 days.

We use the ID token for verifying the user is authenticated, and we do this by passing the token to an internal service that verifies the token hasn't been manipulated by checking it against the AWS [JWK](https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-with-identity-providers.html#amazon-cognito-identity-user-pools-using-id-and-access-tokens-in-web-api) that cryptographically signed the token.

The JWK is a set of keys that are the public equivalent of the private keys used by AWS to digitally sign the tokens. We acquire these via a standard format endpoint:

`https://cognito-idp.{region}.amazonaws.com/{userPoolId}/.well-known/jwks.json.`

## API Limits

One issue we stumbled across recently was the API limits, which meant we couldn't make any further API requests (and for an indeterminate amount of time) 🤔

Seems there is a [Cognito API limits](https://docs.aws.amazon.com/cognito/latest/developerguide/limits.html) reference page, but it's still unclear how long you have to wait before you can start making requests again.

## Which is the right solution?

The answer: it depends.

For me the server-side solution made the most sense, and although difficult in the beginning (primarily due to documentation and general mis-understandings about the difference between User Pools and Identity Pools) we found it worked the best for our requirements, and gave us the most flexibility.

## User Pool Configuration

As far as the User Pool is concerned you'll need a few things:

> Note: this is based on a server-side solution.

- **Application Client**: this will generate a client 'id' and 'secret', which your application(s) will need to use when making certain API calls †  

- **Federated Identity Providers**: this is where you tell Cognito about your social providers (facebook, google etc).  

- **IAM User**: some API calls require AWS credentials (access/secret key), so you'll need to create an IAM user and define the various Cognito APIs you want it to have access to.

> † even if you opt for the 'hosted ui' solution, you'll still need an application client (for two reasons). Firstly you'll configure which 'providers' you want your client app to support, and this will affect what the hosted ui will display to your users. Secondly, the client app id is used as part of the hosted ui uri; meaning you can have _different_ hosted ui's (all configured slightly differently).

## IAM User

The IAM _user_ is necessary as we have to provide some credentials to the [boto client](https://boto3.readthedocs.io/en/latest/reference/services/cognito-idp.html) ([boto](https://boto3.readthedocs.io/en/latest/reference/core/boto3.html) is the Python SDK) in order for it to make certain API calls. Below is an example of the code to instantiate a client:

```
client = boto3.client('cognito-idp', **{'aws_access_key_id': access_key,
                                        'aws_secret_access_key': secret_key,
                                        'region_name': region})
```

Notice the service name is `cognito-idp` and not `cognito-identity`. I mention this as the docs specify two different services "CognitoIdentity" and "CognitoIdentityProvider", which when we were first learning about Cognito we presumed the latter "CognitoIdentityProvider" was something associated with Cognito Identity Pools. 

As we were only interested in the User Pool functionality, we found it strange that the small number of examples we found online all referenced `cognito-idp`. 

So we struggled for a bit to understand the difference, and although we used the "CognitoIdentityProvider" service (i.e. `cognito-idp`), we were confused for a long time as to why that was the case.

Turns out that Cognito's "User Pool" is itself fundamentally a _identity provider_, and because of that you can configure a "Identity Pool" to have a "User Pool" associated within it (along with more common external identity providers such as Facebook and Google).

But as far as the SDK is concerned, you'll always use `cognito-idp` when dealing with a User Pool or an Identity Pool.

## Lambda IAM Role

Below is an example IAM role policy you can use for AWS Lambda (if you're using the hosted ui option and need lambda for logic processing):

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

- **Sign out URL(s)**:  
  `https://auth-api.example.com/auth/signout`

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

## Conclusion

There's so much more to the story, but I think this post is long enough as it is and I don't want to keep you any longer. If you have any questions, then please reach out to me on twitter.

I personally found the documentation around Cognito (and the various tools) to be both overwhelming and underwhelming. Not to mention confusing in places, as well as just downright frustrating at times. 

Hopefully you found this short break down of AWS Cognito useful. There's so much more still to dive into, but this should give you at least a decent starting point.
