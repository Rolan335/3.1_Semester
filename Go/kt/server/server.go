package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nats-io/nats.go"

	_ "github.com/lib/pq"
)

type Client struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Product struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
}

type Order struct {
	Login        string `json:"login"`
	ProductsName string `json:"productsName"`
	Address      string `json:"address"`
}

func main() {
	http.HandleFunc("/registration", registration)
	http.HandleFunc("/order", order)
	http.HandleFunc("/getProducts", getProducts)
	// Connect to a server
	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		panic(err)
	}
}


func order(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	order := Order{}
	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &order)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
	db, _ := sql.Open("postgres", connStr)
	defer db.Close()

	rows, _ := db.Query("SELECT name, description, quantity, price from products where name = $1;", order.ProductsName)
	var row Product
	for rows.Next() {
		var rowSlice Product
		err := rows.Scan(&rowSlice.Name, &rowSlice.Description, &rowSlice.Quantity, &rowSlice.Price)
		if err != nil {
			fmt.Println(err)
		}
		row = rowSlice
	}

	if (row == Product{}) {
		w.Write([]byte("Товара нет"))
	} else {
		hashedAddress := fmt.Sprintf("%x", sha256.Sum256([]byte(order.Address)))
		_, sqlErr := db.Exec(`INSERT INTO order_history (client_login, products_name, address)
	VALUES ($1, $2, $3);`, order.Login, order.ProductsName, hashedAddress)
		if sqlErr != nil {
			fmt.Println(sqlErr)
		}
		_, updErr := db.Exec(`UPDATE products set quantity = quantity - 1 where name = $1;`, order.ProductsName)
		if updErr != nil {
			fmt.Println(sqlErr)
		}
		
		//Куда-то отправляем новый заказ по NATS. Nats сервер поднять не получилось))
		nc, _ := nats.Connect("0.0.0.0:4222")
		jsonRes, _ := json.Marshal(row)
		nc.Publish("new order", jsonRes)

		w.Write([]byte("Ваш заказ отправлен"))
	}
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var productList []Product

	connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
	db, _ := sql.Open("postgres", connStr)
	defer db.Close()

	rows, _ := db.Query(`SELECT NAME, DESCRIPTION, QUANTITY, PRICE FROM PRODUCTS`)
	for rows.Next() {
		var rowSlice Product
		err := rows.Scan(&rowSlice.Name, &rowSlice.Description, &rowSlice.Quantity, &rowSlice.Price)
		if err != nil {
			fmt.Println(err)
		}
		productList = append(productList, rowSlice)
	}
	result, _ := json.Marshal(productList)
	w.Write(result)
}

func registration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	Client := Client{}
	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &Client)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(Client.Password))) //Хэшированный пароль

	//Теперь всё это передаём в бд
	connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
	db, _ := sql.Open("postgres", connStr)
	defer db.Close()

	sqlStatement := `INSERT INTO clients (LOGIN, PASSWORD, EMAIL)
	VALUES ($1, $2, $3);`
	_, sqlErr := db.Exec(sqlStatement, Client.Login, passwordHash, Client.Email)
	if sqlErr != nil {
		fmt.Println(sqlErr)
	}
	w.WriteHeader(200)
}
