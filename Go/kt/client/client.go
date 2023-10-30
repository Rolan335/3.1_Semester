package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Order struct {
	Login        string `json:"login"`
	ProductsName string `json:"productsName"`
	Address      string `json:"address"`
}

type Product struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
}

var serverURL string

func main() {
	serverURL = "http://localhost:8082"

	user := registration()

	fmt.Println("Доступные товары:")
	fmt.Println(getProducts(user))

	fmt.Println("Хотите заказать? Y/N")
	var ans string
	fmt.Scanln(&ans)
	if ans == "Y" {
		createOrder(user)
	} else {
		fmt.Println("Goodbye")
	}
}

func getProducts(user User) []Product {
	data, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(serverURL+"/getProducts", "application/json", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	resData, _ := io.ReadAll(res.Body)
	// json.NewDecoder(res.Body).Decode(&jsonRes)
	var jsonData []Product
	json.Unmarshal(resData, &jsonData)
	return jsonData
}

func createOrder(user User) {
	order := Order{}
	var ans string

	order.Login = user.Login

	fmt.Println("Введите наименование товара")
	fmt.Scanln(&ans)
	order.ProductsName = ans

	fmt.Println("Введите адрес доставки")
	fmt.Scanln(&ans)
	order.Address = ans

	data, _ := json.Marshal(order)
	res, err := http.Post(serverURL+"/order", "application/json", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	resData, _ := io.ReadAll(res.Body)
	fmt.Println(string(resData))
}

func registration() User {
	User := User{}
	fmt.Println("Для начала, зарегистрируйтесь")
	fmt.Println("Введите логин")
	fmt.Scanln(&User.Login)
	fmt.Println("Введите пароль")
	fmt.Scanln(&User.Password)
	fmt.Println("Введите почту")
	fmt.Scanln(&User.Email)
	return User
}
