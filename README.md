# go-service-oriented-package

Instead of putting all the dependencies (database, config etc) into a single repository, separate the services from the dependencies.


In short, this package should only contain the domain application services, with the interfaces defined as an interface. To use and deploy an API using this service, create another package.
