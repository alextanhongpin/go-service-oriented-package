# go-service-oriented-package

Instead of putting all the dependencies (database, config etc) into a single repository, separate the services from the dependencies.


In short, this package should only contain the domain application services, with the interfaces defined as an interface. To use and deploy an API using this service, create another package.



Some design considerations
- the services does not implement transactions - it is up to the caller to implement it externally
- the services are meant to be as generic as possible, it does not cater specific usecases
- the service does not return events etc, it is up to the caller to handle it themselves


## Steps instead of dependencies


## Steps guide

- use command verb
- hint at dependencies, e.g. saveX, queueX, notifyX, emailX
- allow decision making, whenCondition
