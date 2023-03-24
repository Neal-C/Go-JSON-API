package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

//jackc/pgx is to be preferred rather than lib/pq ?

type Storage interface {
	CreateAccount(*Account) error
	GetAccountByNumber(int) (*Account, error)
	GetAccountById(int) (*Account, error)
	GetAccounts() ([]*Account, error)
	UpdateAccount(*Account) error
	DeleteAccount(int) error
}

type PostgresStore struct {
	db *sql.DB
}

//docker run --name some-postgres -e POSTGRES_PASSWORD=gobank -p 5555:5432 -d postgres
//!REMEMBER TO USE ENVIRONMENT/APPLICATION PROPERTIES

func NewPostgresStore()(*PostgresStore, error){
	// db, err := sql.Open("postgres", "postgres://{user}:{password}@{hostname}:{port}/{database-name}?sslmode=disable")

	connStr := "postgres://postgres:gobank@localhost:5555/postgres?sslmode=disable";
	
	db, err := sql.Open("postgres", connStr);
	
	if err != nil {
		defer db.Close();
		return nil, err
	}

//Finally, we are going to call the Ping() method on the sql.DB object we got back from the open function call.

// It is vitally important that you call the Ping() method becuase the sql.Open() function call does not ever create a connection to the database. Instead, it simply validates the arguments provided.

// By calling db.Ping() we force our code to actually open up a connection to the database which will validate whether or not our connection string was 100% correct.

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil;

}


func (selfPostgresStore *PostgresStore) Init() error {
	return selfPostgresStore.createAccountTable();
}

func (selfPostgresStore *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account  (
		id SERIAL PRIMARY KEY UNIQUE,
		first_name VARCHAR(50) NOT NULL,
		last_name VARCHAR(50),
		number SERIAL,
		hashed_password VARCHAR(60) NOT NULL,
		balance INTEGER DEFAULT 0,
		created_at TIMESTAMP
	)`;

	_, err := selfPostgresStore.db.Exec(query);
	return err;
}

func (selfPostgresStore *PostgresStore) CreateAccount(account *Account) error {
	query := `INSERT INTO account (
		first_name, last_name, number, hashed_password, balance, created_at
	) 
	VALUES
	(
		$1,$2,$3,$4,$5,$6
	)`;

	_ , err := selfPostgresStore.db.Query(query,account.FirstName, account.LastName, account.Number, account.HashedPassword, account.Balance, account.CreatedAt);

	if err != nil {
		return err;
	};


	return nil;
}

func (selfPostgresStore *PostgresStore) GetAccountByNumber(number int) (*Account, error){
	query := "SELECT * FROM account WHERE number = $1";
	rows, err := selfPostgresStore.db.Query(query, number);
	if err != nil {
		return nil, err;
	};
	
	for rows.Next(){
		return scanIntoAccount(rows);
	};

	return nil, fmt.Errorf("account [%d] not found 🤔", number);
}

func (selfPostgresStore *PostgresStore) GetAccountById(id int) (*Account, error) {
	query := "SELECT * FROM account WHERE id = $1";
	rows, err := selfPostgresStore.db.Query(query,id);
	if err != nil {
		return nil, err;
	};
	
	for rows.Next(){
		return scanIntoAccount(rows);
	};

	return nil, fmt.Errorf("account %d not found 🤔", id);

}

func (selfPostgresStore *PostgresStore) GetAccounts() ([]*Account, error){

	query := `SELECT * FROM account`;
	rows, err := selfPostgresStore.db.Query(query);
	if err != nil {
		return nil, err;
	}

	var accounts []*Account;

	for rows.Next(){
		account, err := scanIntoAccount(rows);
		//fields must be in order, or there could Undefined Behavior
		//or inconsistent behavior
		if err != nil {
			return nil, err;
		}

		accounts = append(accounts, account);
	}
	return accounts, nil;
}



func (selfPostgresStore *PostgresStore) UpdateAccount(*Account) error {
	return nil;
}


func (selfPostgresStore *PostgresStore) DeleteAccount(id int) error {
	query := "DELETE FROM account WHERE id = $1";
	_, err := selfPostgresStore.db.Query(query, id);

	if err != nil {
		return err
	};
	
	return nil;
}

func scanIntoAccount(rows *sql.Rows) (*Account, error){
	var account = &Account{};
		//fields must be in order, or there could Undefined Behavior
		//or inconsistent behavior
		err := rows.Scan(
			&account.ID, 
			&account.FirstName, 
			&account.LastName,
			&account.Number,
			&account.HashedPassword,
			&account.Balance,
			&account.CreatedAt); 

	return account, err;
}
