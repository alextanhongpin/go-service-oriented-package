# go-service-oriented-package


Unlike clean architecture that focuses on abstracting dependencies, step-driven-development focuses on defininig steps as interface.

When changes happened in the codebase, it will usually be the following:
- new feature added (new steps!)
- removal of a feature
- change in dependencies
- adding/removing/updating a new flow

However, if those changes leads to a lot of changes in other parts of the code such as tests, then the system is brittle.


Goal
- code as documentation
- abstract implementation details
- clearer rules
- ability to make changes real quick (adding or removing does not affect the test)


## Do and Don'ts

- don't return error from abstraction, let the client choose what error to return
- don't use the name usecase - it is wrong. Usecase defines the interaction between user and the system. Within the system, we outline the system flow through steps.
- do prefer a function with a single responsibility and dependencies
- the services does not implement transactions - it is up to the caller to implement it externally
- the services are meant to be as generic as possible, it does not cater specific usecases
- the service does not return events etc, it is up to the caller to handle it themselves
- define the business flows in another package. like domain package, this package should not have external dependencies

## Steps instead of dependencies


## Steps guide

- use command verb
- hint at dependencies
	- saveX for db
	- queueX for message queue
	- notifyX for push notification
	- emailX for sending email
	- smsX for sending sms
	- cacheX for caching something
- allow decision making, whenCondition so that client can return their own error kind
