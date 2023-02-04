package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type UserStorage interface {
	CreateUser(*CreatUserRequest) (string, error)
	GetUsers() ([]*User, error)
	GetUserByEmail(string) (*User, error)
	GetUserById(string) (*User, error)
	DeleteUser(string) error
}

type MongoStorage struct {
	client *mongo.Client
	ctx    context.Context
}

type MongoUserCollection struct {
	store      *MongoStorage
	collection *mongo.Collection
}

func NewMongoStorage() (*MongoStorage, error) {
	client, ctx, err := connect("mongodb://localhost:27017")
	if err != nil {
		return nil, err
	}

	err = ping(client, ctx)
	if err != nil {
		return nil, err
	}

	return &MongoStorage{client, ctx}, nil
}

func NewMongoCollection(m *MongoStorage, database, collection string) *MongoUserCollection {
	userDB := m.client.Database(database)
	return &MongoUserCollection{store: m,
		collection: userDB.Collection(collection)}
}

func connect(uri string) (*mongo.Client, context.Context, error) {

	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	return client, ctx, err
}

func ping(client *mongo.Client, ctx context.Context) error {
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	fmt.Println("Mongo DB connected successfully")

	return nil
}

func close(m *MongoStorage) {
	defer func() {
		if err := m.client.Disconnect(m.ctx); err != nil {
			panic(err)
		}
	}()
}

func (c *MongoUserCollection) insertOne(doc any) (*mongo.InsertOneResult, error) {
	result, err := c.collection.InsertOne(c.store.ctx, doc)

	return result, err
}

func (c *MongoUserCollection) findOne(filter, object any) error {
	result := c.collection.FindOne(c.store.ctx, filter)
	if result.Err() != nil {
		return result.Err()
	}

	return result.Decode(object)
}

func (c *MongoUserCollection) CreateUser(u *CreatUserRequest) (string, error) {
	insertResult, err := c.insertOne(u)
	if err != nil {
		return "", err
	}
	docId := insertResult.InsertedID.(primitive.ObjectID).Hex()

	return docId, err
}

func (c *MongoUserCollection) GetUsers() ([]*User, error) {
	ctx := c.store.ctx
	cursor, err := c.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	users := []*User{}
	for cursor.Next(ctx) {
		user := new(User)
		if err = cursor.Decode(user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (c *MongoUserCollection) GetUserByEmail(email string) (*User, error) {

	filter := bson.M{"email": email}
	user := new(User)
	if err := c.findOne(filter, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (c *MongoUserCollection) GetUserById(idStr string) (*User, error) {
	docId, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, fmt.Errorf("bad user id: %s", idStr)
	}

	filter := bson.M{"_id": docId}
	user := new(User)
	if err := c.findOne(filter, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (c *MongoUserCollection) DeleteUser(id string) error {
	return nil
}

type AccountStorage interface {
	CreateAccount(*Account) error
	DeleteAccount(string) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(string) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "postgres://postgres:bank-api-db@localhost/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("Postgres DB connected successfully")

	return &PostgresStore{db}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	// query := `DROP TABLE IF EXISTS account`
	// _, err := s.db.Exec(query)

	query := `CREATE TABLE IF NOT EXISTS account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		balance float,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)

	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `INSERT INTO account
	(first_name, last_name, number, balance, created_at)
	VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
	)
	return err
}

func (s *PostgresStore) DeleteAccount(id string) error {
	query := `DELETE FROM account WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *PostgresStore) UpdateAccount(acc *Account) error {
	return nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	query := `SELECT * FROM account`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {

		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *PostgresStore) GetAccountByID(id string) (*Account, error) {
	query := `SELECT * FROM account WHERE id = $1`

	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account id# %s not found", id)
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return account, err
}
