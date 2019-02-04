---
title: Resume
description: Staff Software Engineer @BuzzFeed
comments: false
menu: main
weight: -170
---

- [Hello!](#hello)
- [Working Together](#working-together)
- [Connect](#connect)
- [Brief History](#brief-history)
- [Impact!](#impact)
- [Talks](#talks)
- [Interviews](#interviews)
- [Published](#published)
- [Popular articles](#popular-articles)
- [Open-Source](#open-source)
- [Tools, Languages and Tech](#tools-languages-and-tech)
- [Summary](#summary)

<img src="../images/profile.jpg">

## Hello!

Hello, my name is Mark and I work @[BuzzFeed](https://www.buzzfeed.com/) as a Staff Software Engineer (formerly Principal Engineer @[BBCNews](http://www.bbc.co.uk/news)).

I'm also a published author with Apress and LeanPub:

### Apress

- [Quick Clojure: Effective Functional Programming](http://www.apress.com/9781484229514)
- [Pro Vim](http://www.apress.com/9781484202517)
- [Tmux Taster](http://www.apress.com/gb/book/9781484207765)

### LeanPub

- [Python for Programmers](https://leanpub.com/pythonforprogrammers)
- [Programming in Clojure](https://leanpub.com/programming-clojure/)

## Working Together

> "as smart as developers are, they are not always good at explaining things in a way that makes human sense. not you. you are an exception. you are A+".

I ❤️ this feedback. It came from someone I was mentoring who worked in Product Support at BuzzFeed. She was interested in getting a better understanding of how software systems were designed/architected, and also how to know what types of questions she should ask when investigating technical incidents.

Her feedback also hints at something bigger which I strive for: to help others to do their best work and to push/promote the good work other engineers do (especially those from either a diverse background or minority).

I care a lot about the people I work with, and aim to build meaningful relationships with people across the company. In doing so I hope to ensure that we collectively are able to work as a cohesive unit, and thus provide great value to our users.

## Connect

You can find me online at the following locations:

- [integralist.co.uk](http://www.integralist.co.uk/)
- [github.com/integralist](https://github.com/integralist)
- [twitter.com/integralist](http://www.twitter.com/integralist)
- [keybase.io/integralist](https://keybase.io/integralist)
- [linkedin.com/mark-mcdonnell](https://www.linkedin.com/in/mark-mcdonnell-08800565)

## Brief History

### BuzzFeed (June 2016 - present)

The journey has just begun...

BUT the story so far is that I joined as a Senior Software Engineer as part of a new core UK dev team being pushed to take on some important new projects for the company.

Alongside that we were tasked with decommissioning a 10yr+ legacy Perl monolithic application stack over to various Python and Go services.

I've worked primarily within BuzzFeed's OO-INFRA group which sits somewhere in-between traditional infrastructure/operation teams and engineering teams building user facing products. Our motivations were to make the lives of our fellow engineers easier by building tools, services and abstractions that enabled them to work more quickly and efficiently.

In January 2018 I was promoted to Staff Software Engineer, after helping to design/architect, develop and maintain some of BuzzFeed's key infrastructure and software (CDN, caching strategies, routing behaviours, and security/authentication related concerns).

Each year I would also participate in the various working groups and mentoring programs, and become part of the 'on-call' rota and handle interactions with the Hackerone program.

> Note: I'm a remote worker and currently my team is based in New York, so good communication (+ focus and work ethic) is essential.

### BBC (Jan 2013 - June 2016)

I joined [BBC News](http://www.bbc.co.uk/news) as a client-side/mobile specialist within their Core News team. Within the year I had moved into a senior engineering role. The (then) Technical Lead for the BBC News Frameworks team requested I join them in order to help the organisation transition from its current platform over to one built upon the AWS platform.

I started in the Frameworks team building and designing back-end architecture for different microservices hosted upon the AWS platform, and we developed these services primarily using JRuby. In October 2014, I was offered the role of Technical Lead. 

Near the end of 2015 I decided to change roles to Principal Software Engineer, as my previous role required more of my time to be spent in meetings and handling line manager duties, whereas I wanted to focus my time more on helping my team solve technical problems.

### Storm Creative (Feb 2001 - Dec 2012)

I started working at the agency [Storm Creative](http://www.stormcreative.co.uk/) straight out of college. I was always focused on learning and improving my skill set - both technical and communication skills - the latter helped me communicate better with both clients and other stakeholders/colleagues. 

I progressed upwards through the organisation, moving from initially being a client-side web developer (this included doing animations utilising ActionScript 3.0) to becoming a server-side developer (ASP.NET, PHP and Ruby), then onto becoming a Technical Lead across all projects and finally becoming the Digital Media Manager responsible for my own team of four engineers and overseeing all aspects of our projects.

## Impact!

I'd like to share the various things I've worked on over the years and the impact/value this work has provided.

> Note: the large majority of my 'impact' has been as a _remote_ worker. My hope is that the following list demonstrates how I've successfully made a positive impact (both as an IC and as a Team Lead) while also being 100% remote.

### 2018

- **What**: I replaced BuzzFeed's use of NGINX+ (a very expensive commercial product that was being used as part of a critical piece of BuzzFeed's infrastructure) with the open-source equivalent.  
  **Why**: This was a [hack week project](../../pdfs/hackweek_2018_nginx.pdf). It took one day to implement the changes, one day to test and verify behaviours in a staging environment, followed by a quick rollout to production.   
  **Impact**: This saved the organization $60,000 a year in licensing fees.

- **What**: Designed and co-implemented new authentication system built in Python on top of AWS Cognito.   
  **Why**: Decommission our legacy authentication system which was tightly coupled to a 10yr+ monolithic Perl application.  
  **Impact**: Enabled more services to offer authentication, thus allowing more community driven features across our products.

- **What**: Built a Python package that wraps scrypt.  
  **Why**: Provide a consistent interface when requiring a hashing function.  
  **Impact**: Engineers unfamiliar with the particulars of various security protocols (e.g. the various hashing mechanisms or the difference to encryption) could safely utilise hashing without having to understand the implementation.

- **What**: Helped promote the benefits of Kim Scott's 'Radical Candor', Marshall Rosenberg's 'Nonviolent Communication' and Fred Kofman's 'Authentic Communication' to various teams across BuzzFeed.   
  **Why**: Effective, clear and compassionate communication benefits everyone.  
  **Impact**: Teams were becoming more productive as the confidence to give the most appropriate and direct feedback necessary to catch both interpersonal issues and team concerns happened much more quickly.

- **What**: Introduced Wednesday lunch videos/presentations.   
  **Why**: To motivate and inspire our development teams.  
  **Impact**: People had fun listening to interesting topics (not all tech related), and having a source of conversation and discussion beyond the lunch hour and in some cases helped to inspire changes that fed back into the company.

- **What**: Designed and implemented a Python Tornado web handler decorator responsible for acquiring/caching/revalidating an asymmetrical public key.   
  **Why**: To protect services from unauthenticated access (the public key was used to sign JWTs provided by an authentication proxy we had built previously in Go).  
  **Impact**: Helped engineers to quickly integrate with our custom built authentication service and provide a consistent experience across the platform.

- **What**: Co-designed and co-implemented a Go based reverse proxy acting as an authentication layer in front of BuzzFeed services.   
  **Why**: Part of a plan to decommission our legacy authentication system.  
  **Impact**: The use of JWTs helped to implement a stateless system for providing authenticated access to services, thus making the system easier to reason about, and enabled teams to decouple themselves from our legacy Perl stack.

### 2017

- **What**: Implemented README validator in Python.   
  **Why**: As part of BuzzFeed's "Better Docs" initiative (of which I was a core member of its Working Group).  
  **Impact**: This helped BuzzFeed to track the success of its new "Doc Day" event, which supports staff across the org in reviewing and improving software documentation.

- **What**: Led the effort to document, improve, and educate others on the state of BuzzFeed's monitoring.   
  **Why**: Our monitoring system was very noisy, which made going on-call a much more stressful proposition.  
  **Impact**: . I also wrote a [community blog post](/posts/monitoring-best-practices/) sharing and explaining a lot of what we did, along with sharing [a template Runbook](https://docs.google.com/document/d/1eaT9KMam5zq7lT-5OVz9T91RJQUx-qx2q6WnKSvxC_U/edit?usp=sharing) for operational safety.

- **What**: Core member of the BuzzFeed “Better-Docs” Working Group.   
  **Why**: We aim to improve documentation and its discoverability for BuzzFeed Tech.  
  **Impact**: We standardized the majority of doc formats, the creation and maintenance of doc tooling, and continued to educate ourselves and the BF Tech community about the importance of good documentation.

- **What**: Tech Lead for the Site Infra 'Resilience' team.   
  **Why**: Necessary to help improve the stablity and resilience of BuzzFeed's existing services while helping to educate development teams on the various best practices.  
  **Impact**: We designed a disaster recovery strategy specific for BuzzFeed's needs (called 'Plan Z') which helped to facilitate multiple failure scenarios and failovers for many of our service providers (alongside that primary task we helped improve the resilience for many BuzzFeed services).

- **What**: Built an operations Slackbot in Go.   
  **Why**: This was implemented as part of BuzzFeed's 'Hack Week' but was quickly put into general rotation as is still used for all service incidents.  
  **Impact**: Enabled all BuzzFeed staff (whether technical or not) to quickly spin up either a public or private incident channel in Slack, while allowing interested parties to be auto-invited based upon an emoji reaction implementation. The tool also allowed people to search for operational runbooks stored within our organizations Google Drive.

- **What**: Designed and implemented a round-robin, multi-cloud provider nginx solution for serving static assets.   
  **Why**: To help provide greater resilience when serving client-side static assets such as images or scripts.  
  **Impact**: The tooling we built around this implementation helped to make the process of deploying and serving static assets efficiently much easier.

- **What**: Technical Lead and architect for a dynamic video player service.   
  **Why**: To enable asynchronous editor workflows.  
  **Impact**: Enabled flexible video selection for end users, while helping to promote BuzzFeed's own brand of video content outside of YouTube (which would otherwise require us to lose potential profit).

- **What**: Designed and implemented [a Go CLI tool for deploying Fastly VCL changes](https://github.com/integralist/go-fastly-cli).   
  **Why**: The existing process for deploying Fastly VCL was manual and time consuming, and prone to mistakes.  
  **Impact**: Helped unblock engineers who needed a more efficient way to rollout changes, while allowing them to diff and validate changes locally without having to sign-in to Fastly's otherwise confusing UI.

- **What**: Refactored existing HTTP Cache Client Python package.   
  **Why**: Original design was a facade around a multi-tiered cache abstraction over a Python HTTP client. This proved to be too limiting for engineers.  
  **Impact**: Utilized an Adapter pattern internally in order to provide a unified interface, thus making it easier for various HTTP clients to be provided instead of locking the caller down to a single client type.

- **What**: Implemented GitHub hook mechanism for detecting API changes and generating updated documentation.   
  **Why**: Documentation would often go stale because engineers would make changes but not re-run the rendering tools to generate new docs.  
  **Impact**: Enabled engineers to make changes without having to think about generating new documentation or having to know how to use the various tools for generating documentation.

- **What**: Refactored legacy VCL code and spent time building necessary abstractions.   
  **Why**: Original code was difficult to understand and meant only a blessed few engineers understood how it all worked.  
  **Impact**: Opened up the CDN to more engineers and helped to provide abstractions (such as for logging) to make working with VCL easier for those new to the language.

### 2016

- **What**: Migrated Fastly’s version of varnish/vcl 2.x to standard 4.1.   
  **Why**: Support switching to an alternative backup CDN.  
  **Impact**: Strengthened our relationship with Site Reliability, while also building confidence in a failover CDN.

- **What**: Designed and implemented generic GitHub Pull Request template file.   
  **Why**: Consistency and standardization of how pull requests are structured. The final format was based loosely on an [old blog post](/posts/github-pull-request-formatting/) I wrote (back before GitHub offered their template feature).  
  **Impact**: Clearer problem/solution descriptions that enabled engineers not familiar with the services to understand the changes being proposed.

- **What**: Implemented a smoke test scheduler service in Python.   
  **Why**: Catch regressions with BuzzFeed's primary routing service.  
  **Impact**: Helped engineers to identify integration problems where routing changes would have adverse unexpected effects.

- **What**: Led development across a mostly US based team, and the rollout of a new critical routing service.   
  **Why**: The routing behaviours for BuzzFeed were locked down to those blessed few who understood the CDN and VCL.  
  **Impact**: Enabled the entire engineering department to make routing changes based on complex sets of dynamic input and requirements via a simple config driven workflow.

- **What**: Porting of Perl services over to Python [BFF](http://samnewman.io/patterns/architectural/bff/) services.   
  **Why**: Decommission of 10yr+ monolithic Perl application.  
  **Impact**: Increased BuzzFeed's recruitment opportunities by expanding the amount of services written in Python (compared to hiring Perl developers), as well as improving the code quality of those services migrated.

- **What**: Proposed usage of specific Python linters and related tooling.   
  **Why**: Code consistency and easier debugging of code.  
  **Impact**: Improved overall code quality.

- **What**: Defined "[The Perfect Developer Qualities](https://gist.github.com/Integralist/3f8089345a1236b374a7a5b8a13591a1)".   
  **Why**: To inspire and motivate my colleagues.  
  **Impact**: Engineers from across the organization reached out to me to share their thoughts, feedback and general appreciation for the time and consideration (as well as the obvious past experience) that led to this ideal list of character traits.

- **What**: Released the open-source project [go-elasticache](https://github.com/Integralist/go-elasticache).   
  **Why**: Share useful tools that would benefit others.  
  **Impact**: Improved the developer experience when working with AWS's ElastiCache service.

- **What**: Led performance testing, analysis and resolution of scaling issues for the BBC's internal "Mozart" platform (written in Ruby).   
  **Why**: Network bottlenecks were causing issues during load testing.  
  **Impact**: Helped to identify specific service within the overall architecture that resulted in it being rewritten in Go and thus resolving the scaling performance issues.

- **What**: Implemented simple, yet performant, URL monitoring system in Bash called [Bash Watchtower](/posts/bash-watchtower/).   
  **Why**: Previous version was a complicated and over engineered Node application (it was a colleagues pet project, and no one in the organization used Node at the time). It was also laden with NPM packages which made installing and running a very slow process.  
  **Impact**: Improved deployment speed, helped other engineers understand the code base by using a language they were more familiar with, and simplified the overall code.

- **What**: Created and led BBC News "Coding and Architecture" working group.   
  **Why**: We were charged with ensuring best practices were adhered to.  
  **Impact**: Improved the overall quality of new services being developed, and helped us to communicate with a wider range of the organization.

- **What**: Co-designed and co-implemented the BBC News "Mozart" platform.   
  **Why**: Simplify the ability to build up dynamic page composition.  
  **Impact**: Enabled teams to more easily build up complex pages of individual components. It also helped path the way for the organization to move away from internal hosted system to the AWS platform, while enabling developers to utilize easier languages and tools.

### 2015

- **What**: Represented BBC at AWS' week long re:Invent technical conference in Las Vegas.   
  **Why**: To learn more about the new AWS services that could benefit the organization.  
  **Impact**: Networking with lots of different companies and helping to promote the work that the BBC does (specifically the engineering arm of the organization).

- **What**: Co-designed and co-implemented a Go based CLI tool called "Apollo".   
  **Why**: Abstract away certificate based authentication to internal APIs.  
  **Impact**: Enabled teams to more easily deploy services to the AWS platform.

- **What**: Team Lead for BBC News Frameworks team.   
  **Why**: To help my team grow and to learn.  
  **Impact**: Helped to promote a large segment of my team into senior position roles.

- **What**: Won "Connecting the News" Hack Day event.   
  **Why**: Event for different news organizations to come together around a shared data source (provided by the BBC) and to see what interesting tools and services can enhance that data.  
  **Impact**: Networking with engineering teams across different news platforms helped to inform potential ideas for our own services. Showcased BBC News as a great place to work.

- **What**: Released BBC Newsbeat v2.   
  **Why**: First fully AWS product from BBC News.  
  **Impact**: Started the movement of services from using an internal hosting platform onto the AWS platform.

- **What**: Tech Lead for General Elections.   
  **Why**: The General Elections was a big event for BBC News.  
  **Impact**: Successful build, deploy and monitoring of election reporting platform.

- **What**: Rebuilt and migrated BBC's Market Data to AWS using the BBC's open-source Alephant framework, of which I was a co-author.   
  **Why**: Fix an old and un-maintained, yet critical, data service.  
  **Impact**: Modernized and improved this essential financial market service for its stakeholders and enabled further extension by other engineering teams.

### 2014

- **What**: Designed and implemented "Jello" which was an internal synchroniation service between Trello and Jira.   
  **Why**: Teams preferred to use Trello, while the rest of the organization was using a very old version of Jira.  
  **Impact**: Enabled teams to benefit from the speed and feature set of Trello without having to manually track tasks back into Jira for the rest of the organizations visibility.

- **What**: Won "Most innovative use of Technology" BBC News Award (Docker CI).   
  **Why**: Legacy Jenkins CI was locked down to centralized operations team.  
  **Impact**: Enabled teams to build and deploy software using any langage or platform supported by Docker.

- **What**: Won "Best Public Relations of the Year" BBC News Award (Pro Vim).   
  **Why**: I like writing and sharing information that helps people be more proficient with the tools they use.  
  **Impact**: Book was well received and opened the Vim editor to wider range of engineers.

- **What**: Co-designed and co-implemented cloud based distributed load testing tool.   
  **Why**: Existing solutions weren't able to scale with our platform.  
  **Impact**: .

- **What**: Organized public speaking event with [Sandi Metz](http://www.sandimetz.com/).   
  **Why**: To build an engineering network event for the London tech community.  
  **Impact**: London tech community got to see an otherwise often unseen internal look at BBC engineering talent and were able to discuss topics of interest.

### 2013

- **What**: Voted "Developer of the Year" at the BBC News awards.   
  **Why**: I had made sure to reach out and affect in a positive way every single aspect of the business and to make a real difference to the developer community within the BBC.  
  **Impact**: A genuine sense of pride that I was able to achieve what I set out to do: make a difference.

- **What**: Led development of the BBC News 'responsive navigation' redesign.   
  **Why**: Part of the new BBC UX rebranding.  
  **Impact**: Resulted in communication with product, design and engineering teams across the entire breadth of the BBC platform. Leading to a new responsive nagivation that was able to successfully accommodate all perspectives and requirements.

- **What**: Invited to [speak at Mozilla offices in Paris](https://speakerdeck.com/integralist/bbc-news-responsive-images).   
  **Why**: To discuss the BBC News responsive images technique to browser vendors such as Apple, Microsoft, Opera, Mozilla and Google.  
  **Impact**: I was able to establish myself as a person of interest to this organizations and an expert in the field when it came to client-side development.

- **What**: Implemented new BBC UX framework.   
  **Why**: The BBC brand was undergoing a organization wide redesign.  
  **Impact**: This was a very long and deliberate implementation and rollout process that helps re-establish BBC News as a leader in the responsive mobile development space and helped showcase BBC News engineering talents.

- **What**: Implemented new BBC [responsive images solution](https://github.com/BBC-News/Imager.js/).   
  **Why**: Scalable and responsive images was not widely supported by browsers with native APIs, meaning custom solutions needed to be implemented.  
  **Impact**: [Public BBC News post](http://responsivenews.co.uk/post/58244240772/imagerjs) proposed our solution to the then difficult problem of how best to serve images in a scalable way to browsers and mobile devices.

- **What**: Introduced the use of [GruntJS](http://gruntjs.com/).   
  **Why**: Ruby and Rake was being used although majority of engineers were unfamiliar with the language and were afraid to make changes or to build new tasks.  
  **Impact**: Improved the ability of engineers to automate project tasks using JavaScript.

- **What**: Member of the [BBC's GEL Responsive Working Group](http://www.bbc.co.uk/gel/).   
  **Why**: To help ensure engineers perspective on how best to implement new UX designs were accounted for.  
  **Impact**: Simplified specific aspects of GELs design.

## Talks

- [Site Router (Video)](https://www.youtube.com/watch?v=md4de3RyN-8): 80 minute presentation on BuzzFeed HTTP routing service abstraction.

- [BBC Talks (Slides)](https://slides.com/markmcdonnell/): various presentations I gave while at the BBC.

- [Imager.js (Slides)](https://speakerdeck.com/integralist/bbc-news-responsive-images): Talk I gave at the Mozilla offices in Paris (which included speakers from: Google, Apple, Microsoft, Adobe, Opera, W3C and Akamai).

## Interviews

- InfoQ: [How BuzzFeed Migrated from a Perl Monolith to Go and Python Microservices](https://www.infoq.com/articles/buzzfeed-microservices-migration)

## Published

I'm a print published and self-published author; I'm also a tech reviewer and am a published author for many popular online organisations (you'll find many technical articles on my own website as well):

### Apress

- [Pro Vim](http://www.apress.com/9781484202517) (Nov 2014)
- [Tmux Taster](http://www.apress.com/gb/book/9781484207765) (Nov 2014)
- [Quick Clojure: Effective Functional Programming](http://www.apress.com/9781484229514) (August 2017)

### Packt

- Tech Reviewer [Grunt Cookbook](https://www.packtpub.com/web-development/grunt-cookbook) (May 2014)
- Tech Reviewer "Troubleshooting Docker" (May 2015)

### LeanPub

- [Programming in Clojure](https://leanpub.com/programming-clojure/) (Jul 2015)
- [Python for Programmers](https://leanpub.com/pythonforprogrammers) (Jun 2016)

### NET Magazine

- [8 ways to improve your grunt set-up](http://www.creativebloq.com/tutorial/8-ways-improve-your-grunt-set-111413407) (Nov 2014) ([PDF](https://dl.dropboxusercontent.com/u/3687270/NetMag%20-%20Grunt.pdf))
- [DalekJS vs CasperJS](https://dl.dropboxusercontent.com/u/3687270/NetMag%20-%20Dalek%20vs%20Casper.pdf) (Nov 2013)

### Smashing Magazine

- [My author page](http://coding.smashingmagazine.com/author/mark-mcdonnell/)
- [Building Software with Make](http://www.smashingmagazine.com/2015/10/building-web-applications-with-make/)
- [How To Build A CLI Tool With Node.js And PhantomJS](http://coding.smashingmagazine.com/2014/02/12/build-cli-tool-nodejs-phantomjs/)
- [How To Build A Ruby Gem With Bundler, TDD, Travis CI & Coveralls, Oh My!](https://www.smashingmagazine.com/2014/04/how-to-build-a-ruby-gem-with-bundler-test-driven-development-travis-ci-and-coveralls-oh-my/)

### NetTuts

- [My author page](http://tutsplus.com/authors/mark-macdonnell)
- [Testing Your Ruby Code With Guard, RSpec & Pry (Part 1 - Ruby/Guard/RSpec)](http://code.tutsplus.com/tutorials/testing-your-ruby-code-with-guard-rspec-pry--cms-19974)
- [Testing Your Ruby Code With Guard, RSpec & Pry (Part 2 - RSpec/Pry/Travis-CI)](http://code.tutsplus.com/tutorials/testing-your-ruby-code-with-guard-rspec-pry-part-2--cms-20290)

### BuzzFeed Tech

- I wrote a three part series on BuzzFeed's core HTTP routing service (built upon NGINX+) called "Scalable Request Handling: An Odyssey":
  - [Part 1](https://tech.buzzfeed.com/scalable-request-handling-an-odyssey-part-1-d91a295af4d8)
  - [Part 2](https://tech.buzzfeed.com/scalable-request-handling-an-odyssey-part-2-ad2433b2f6ed)
  - [Part 3](https://tech.buzzfeed.com/scalable-request-handling-an-odyssey-part-3-c29aac9c39a)

### InfoQ

- Interview: [How BuzzFeed Migrated from a Perl Monolith to Go and Python Microservices](https://www.infoq.com/articles/buzzfeed-microservices-migration)

## Popular articles

The following links are to some of my more 'popular' articles. My main focus when writing is to take a complicated or confusing topic and attempt to distil it, in order for the subject to be more easily understood.

- [Algorithmic Complexity in Python](/posts/algorithmic-complexity-in-python/) (2019)
- [Data Types and Data Structures](/posts/data-types-and-data-structures/) (2019)
- [Engineer to Manager](/posts/engineer-to-manager/) (2018)
- [Interview Techniques](/posts/architecture-interview/) (2018)
- [Post Mortems](/posts/post-mortem-template/) (2018)
- [Thinking about Interfaces in Go](/posts/go-interfaces/) (2018)
- [Multigrain Services](/posts/multigrain-services/) (2018)
- [Authentication with AWS Cognito](/posts/cognito/) (2018)
- [A guide to effective 1:1 meetings](/posts/1-1/) (2018)
- [Project Management in Five Minutes](/posts/project-management-in-five-minutes/) (2018)
- [Interview Topics](/posts/questions-when-interviewing/) (2018)
- [Hashing, Encryption and Encoding](/posts/hashing-and-encryption/) (2018)
- [Computers 101: terminals, kernels and shells](/posts/terminal-shell/) (2018)
- [Statistics and Graphs: The Basics](/posts/statistic-basics/) (2017)
- [Observability and Monitoring Best Practices](/posts/monitoring-best-practices/) (2017)
- [Logging 101](/posts/logging-101/) (2017)
- [Fastly Varnish](/posts/fastly-varnish/) (2017)
- [Profiling Go](/posts/profiling-go/) (2017)
- [Profiling Python](/posts/profiling-python/) (2017)
- [Bits Explained (inc. base numbers, ips, cidrs and more)](/posts/bits-and-bytes/) (2016)
- [Terminal Debugging Utilities](/posts/terminal-debugging-utilities/) (2016)
- [Big O for Beginners](/posts/big-o-for-beginners/) (2016)
- [Git Merge Strategies](/posts/git-merge-strategies/) (2016)
- [HTTP/2](/posts/http2/) (2015)
- [Client Cert Authentication](/posts/client-cert-authentication/) (2015)
- [DNS 101](/posts/dns-101/) (2015)
- [Security basics with GPG, OpenSSH, OpenSSL and Keybase](/posts/security-basics/) (2015)
- [Setting up nginx with Docker](/posts/setting-up-nginx-with-docker/) (2015)
- [Building Software with Make](/posts/building-systems-with-make/) (2015)
- [Thread Safe Concurrency](/posts/thread-safe-concurrency/) (2014)
- [GitHub Workflow](/posts/github-workflow/) (2014)
- [Understanding recursion in functional JavaScript programming](/posts/functional-recursive-javascript-programming/) (2014)
- [Refactoring Techniques](/posts/refactoring-techniques/) (2013)
- [MVCP: Model, View, Controller, Presenter](/posts/mvcp/) (2013)
- [Basic Shell Scripting](/posts/basic-shell-scripting/) (2013)
- [Object-Oriented Design (OOD)](/posts/object-oriented-design/) (2013)
- [Git Tips](/posts/git-tips/) (2012)
- [JavaScript 101](/posts/javascript-101/) (2012)

## Open-Source

> Note: listed alphabetically (i.e. not 'prioritized' in any way)

- [BBC Alephant](https://github.com/BBC-News/alephant):  
The Alephant framework is a collection of isolated Ruby gems, which interconnect to offer powerful message passing functionality built up around the "Broker" pattern.  

- [BBC Imager.js](https://github.com/BBC-News/Imager.js):  
Responsive images while we wait for srcset to finish cooking  

- [Bash Headers](https://github.com/Integralist/Bash-Headers):  
CLI tool, written in Bash script, for sorting and filtering HTTP Response Headers  

- [DOMReady](https://github.com/Integralist/DOMready):  
Cross browser 'DOM ready' function  

- [Go ElastiCache](https://github.com/Integralist/go-elasticache):  
Thin abstraction over the Memcache client package [gomemcache](https://github.com/bradfitz/gomemcache) allowing it to support AWS ElastiCache cluster nodes  

- [Go Fastly CLI](https://github.com/Integralist/go-fastly-cli):  
CLI tool, built in Go, for interacting with the Fastly API  

- [Go Find Root](https://github.com/Integralist/go-findroot):  
Locate the root directory of a project using Git via the command line  

- [Go Requester](https://github.com/Integralist/Go-Requester):  
HTTP service that accepts a collection of "components", fans-out requests and returns aggregated content  

- [Go Reverse Proxy](https://github.com/Integralist/go-reverse-proxy):  
A configuration-driven reverse proxy written in Go (no dependencies outside of the standard library).  

- [Grunt Boilerplate](https://github.com/Integralist/Grunt-Boilerplate):  
Original Grunt Boilerplate  

- [Image Slider](https://github.com/Integralist/HTML5-Image-Slider-Game):  
HTML5 Canvas Game  

- [MVCP](https://github.com/Integralist/MVCP):  
MVC + 'Presenter' pattern in Ruby  

- [Sinderella](https://github.com/Integralist/Sinderella):  
Ruby gem for transforming data object for specified time frame  

- [Spurious Clojure AWS SDK Helper](https://github.com/Integralist/spurious-clojure-aws-sdk-helper):  
Helper for configuring the AWS SDK to use [Spurious](https://github.com/spurious-io/spurious)  

- [Squirrel](https://github.com/Integralist/Squirrel):  
PhantomJS script to automate Application Cache manifest file generation  

- [Stark](https://github.com/Integralist/Stark):  
Node Build Script for serving HTML components

## Tools, Languages and Tech

I don't profess mastery, but I'm adept with most of the below, and I have an aptitude towards learning what I need to get the job done right:

- AWS CloudFormation
- CSS
- Clojure
- [Design Patterns](https://github.com/Integralist/Ruby-Design-Patterns)
- Docker
- [Functional Programming](/posts/functional-recursive-javascript-programming/)
- Git
- Go
- HTML
- JRuby/MRI Ruby
- JavaScript (client-side)
- Jenkins
- Jira
- Make
- [Meta Programming](https://gist.github.com/Integralist/a29212a8eb10bc8154b7)
- Node
- NGINX
- PHP
- Python
- [Refactoring](/posts/refactoring-techniques/)
- Regular Expressions
- Sass
- [Shell Scripting](/posts/basic-shell-scripting/)
- Terraform
- Tmux
- Trello
- Vagrant
- Varnish
- VCL
- Vim

## Summary

I ideally want to get across two fundamental things about me:

1. I'm very passionate about programming and the openness of the web
2. I love getting the chance to learn and experience new things
