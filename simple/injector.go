//go:build wireinject
//+build wireinject

package simple

import "github.com/google/wire"

func IntializeService(isError bool) (*SimpleService, error) {
	wire.Build(
		NewSimpleRepository, NewSimpleService,
		)
	return nil, nil
}


func IntializeDatabaseRepository() *DatabaseRepository {
	wire.Build(NewDatabasePostgreSQL, NewDatabaseMongoDB, NewDatabaseRepository)
	return nil
}

var (
	fooSet = wire.NewSet(NewFooRepository, NewFooService)
	barSet = wire.NewSet(NewBarRepository, NewBarService)
)

func InitializeFooBarService() *FooBarService {
	wire.Build(fooSet, barSet, NewFooBarService)
	return nil
}

var helloSet = wire.NewSet(NewSayHelloImpl, wire.Bind(new(SayHello), new(*SayHelloImpl)))

func InitializeHelloService() *HelloService {
	wire.Build(helloSet, NewHelloService)
	return nil
}