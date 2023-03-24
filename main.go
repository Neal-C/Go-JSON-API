package main

import (
	"flag"
	"fmt"
	"log"
)

//Everything will be package main, so we can scale up whenever needed and not worry about introducing circular dependencies with mishandled modules

func seedAccount(store Storage, name string, lastName string, password string){
	account, err := NewAccount(name, lastName, password);

	if err != nil {
		log.Fatal(err);
	}

	if err := store.CreateAccount(account); err != nil {
		log.Fatal(err);
	}

	fmt.Println("new account =>", account.Number);
}

func seedAccounts(store Storage){
	seedAccount(store, "fart junior", "farter", "fart" );
}

func main(){

	//go run . --seed true;
	seed := flag.Bool("seed", false, "seed the database");
	flag.Parse();
	
	store, err := NewPostgresStore();
	if err != nil {
		fmt.Println("problem");
		log.Fatal(err);
	}

	if err := store.Init(); err != nil {
		log.Fatal(err);
	}

	//seed
	if *seed {
		fmt.Println("seeding the databse")
		seedAccounts(store);
	}

	fmt.Printf("%+v \n", store)
	server := NewAPIServer(":3000", store);

	// WriteTimeout: 15 * time.Second,
	// ReadTimeout:  15 * time.Second,

	server.Run();

}