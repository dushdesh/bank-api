package main

import "log"

func main() {
	mongoStore, err := NewMongoStorage()
	if err != nil {
		log.Fatal("Mongo DB connection failed: ", err)
	}
	defer close(mongoStore)

	
	userStore := NewMongoCollection(mongoStore, "bank-api", "users")
	userService := NewUserService(userStore)

	accountStore, err := NewPostgresStore()
	if err != nil {
		log.Fatal("PG DB connection failed: ", err)
	}

	if err = accountStore.Init(); err != nil {
		log.Fatal("PG table cannot be created: ", err)
	}

	server := NewApiServer(":3000", accountStore, userService)
	server.Run()
}
