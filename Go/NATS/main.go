package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/nats-io/nats.go"

	_ "github.com/lib/pq"
)

var usersMap = make(map[string]User)
var userCounter = 0

type DisplayTable struct {
	Diagonal   float32 `json:"diagonal"`
	Resolution string  `json:"resolution"`
	Type       string  `json:"type"`
	GSync      bool    `json:"gsync"`
}

type MonitorTable struct {
	DisplayID    int  `json:"displayID"`
	GSyncPremium bool `json:"gSyncPremium"`
	IsCurved     bool `json:"isCurved"`
}

type User struct {
	ID_User       int
	Username      string
	Password_Hash string
	Email         string
	Is_Admin      bool
}

type UserRegistration struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// {
//     "username": "rolan",
//     "password": "1111",
//     "email": "abc@mail.ru"
// }

var natsURL string

func main() {
	natsURL = "nats://95.165.107.100:4222"
	// Connect to a server

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	http.HandleFunc("/registration", registration)
	http.HandleFunc("/login", login)

	http.HandleFunc("/addDisplay", addDisplayHandler)
	http.HandleFunc("/addMonitor", addMonitorHandler)

	http.HandleFunc("/allDisplays", allDisplaysHandler)
	http.HandleFunc("/allMonitors", allMonitorsHandler)
	http.HandleFunc("/getDisplayById", getDisplayById)

	http.HandleFunc("/deleteMonitor", deleteMonitorHandler)
	http.HandleFunc("/deleteDisplay", deleteDisplayHandler)

	http.HandleFunc("/makeAdmin", makeAdmin)

	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		nc.Publish("log", []byte("http server error"))
		panic(err)
	}
	for {

	}
}

func getDisplayById(w http.ResponseWriter, r *http.Request) {
	// Connect to a server
	natsURL := "nats://95.165.107.100:4222"
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	token := r.Header.Get("Authorization")
	if _, ok := usersMap[token]; ok {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")

		id := make(map[string]int)

		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()

		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		rows, _ := db.Query("SELECT diagonal, resolution, type, gsync FROM displays where id_displays = $1;", id["id"])
		var displays []DisplayTable
		for rows.Next() {
			var rowSlice DisplayTable
			err := rows.Scan(&rowSlice.Diagonal, &rowSlice.Resolution, &rowSlice.Type, &rowSlice.GSync)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(rowSlice)
			displays = append(displays, rowSlice)
		}
		result, _ := json.Marshal(displays)
		w.Write(result)
		fmt.Println(displays)

		//Логирование
		nc.Publish("log", []byte("command: getDisplayById, userId: "+strconv.Itoa(usersMap[token].ID_User)+" displayID: "+strconv.Itoa(id["id"])))
	} else {
		nc.Publish("log", []byte("Error. User not found. Token: "+token))
		w.WriteHeader(403)
	}
}

func deleteDisplayHandler(w http.ResponseWriter, r *http.Request) {
	// Connect to a server

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	token := r.Header.Get("Authorization")
	if usersMap[token].Is_Admin {
		id := make(map[string]int)
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()

		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		sqlStatement := `DELETE FROM displays WHERE id_displays = $1`
		_, sqlErr := db.Exec(sqlStatement, id["id"])
		if sqlErr != nil {
			fmt.Println(sqlErr)
		}
		nc.Publish("log", []byte("command: deleteDisplay, userId: "+strconv.Itoa(usersMap[token].ID_User)+" displayID: "+strconv.Itoa(id["id"])))
		w.WriteHeader(200)
	} else {
		nc.Publish("log", []byte("Error. Admin not found. Token: "+token))
		w.WriteHeader(403)
	}
}

// TODO Сделать проверку
func deleteMonitorHandler(w http.ResponseWriter, r *http.Request) {
	// Connect to a server

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	token := r.Header.Get("Authorization")
	if usersMap[token].Is_Admin {
		id := make(map[string]int)
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()

		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		sqlStatement := `DELETE FROM monitors WHERE id_monitors = $1`
		_, sqlErr := db.Exec(sqlStatement, id["id"])
		if sqlErr != nil {
			fmt.Println(sqlErr)
		}
		nc.Publish("log", []byte("command: deleteMonitor, userId: "+strconv.Itoa(usersMap[token].ID_User)+" monitorID: "+strconv.Itoa(id["id"])))
		w.WriteHeader(200)
	} else {
		nc.Publish("log", []byte("Error. Admin not found. Token: "+token))
		w.WriteHeader(403)
	}
}

// TODO : Сделать ручку добавления админа
func makeAdmin(w http.ResponseWriter, r *http.Request) {
	// token := r.Header.Get("Authorization")
}

func allMonitorsHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Connecting...")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	token := r.Header.Get("Authorization")
	if _, ok := usersMap[token]; ok {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()
		rows, _ := db.Query("SELECT display_id, gsync_premium, curved FROM monitors;")
		var monitors []MonitorTable
		for rows.Next() {
			var rowSlice MonitorTable
			err := rows.Scan(&rowSlice.DisplayID, &rowSlice.GSyncPremium, &rowSlice.IsCurved)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(rowSlice)
			monitors = append(monitors, rowSlice)
		}
		result, _ := json.Marshal(monitors)
		w.Write(result)
		fmt.Println(monitors)
		nc.Publish("log", []byte("command: allMonitors, userId: "+strconv.Itoa(usersMap[token].ID_User)))
	} else {
		nc.Publish("log", []byte("Error. User not found. Token: "+token))
		w.WriteHeader(403)
	}

}

func allDisplaysHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Connecting...")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	token := r.Header.Get("Authorization")
	if _, ok := usersMap[token]; ok {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()
		rows, _ := db.Query("SELECT diagonal, resolution, type, gsync FROM displays;")
		var displays []DisplayTable
		for rows.Next() {
			var rowSlice DisplayTable
			err := rows.Scan(&rowSlice.Diagonal, &rowSlice.Resolution, &rowSlice.Type, &rowSlice.GSync)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(rowSlice)
			displays = append(displays, rowSlice)
		}
		result, _ := json.Marshal(displays)
		w.Write(result)
		fmt.Println(displays)
		nc.Publish("log", []byte("command: allDisplays, userId: "+strconv.Itoa(usersMap[token].ID_User)))
	} else {
		nc.Publish("log", []byte("Error. User not found. Token: "+token))
		w.WriteHeader(403)
	}
}

// Проверяем, зарегистрирован ли пользователь по нашей мапе
// Если да, отправляем клиенту его токен.
func login(w http.ResponseWriter, r *http.Request) {

	log.Println("Connecting...")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	TempUser := UserLogin{}
	body, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(body, &TempUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Password)))
	token := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Username+passwordHash)))
	stringToken := fmt.Sprintf("%x", token)
	_, ok := usersMap[stringToken]
	if ok {
		w.Write([]byte(stringToken))
		nc.Publish("log", []byte("command: login, userId: "+strconv.Itoa(usersMap[stringToken].ID_User)+" username: "+TempUser.Username))
	} else {
		nc.Publish("log", []byte("Error. User not found. Token: "+token))
		w.WriteHeader(403)
	}
}

func registration(w http.ResponseWriter, r *http.Request) {

	log.Println("Connecting...")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	TempUser := UserRegistration{}
	body, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(body, &TempUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Password))) //Хэшированный пароль

	//Генерируем токен и кидаем в мапу
	userCounter += 1
	token := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Username+passwordHash)))
	usersMap[token] = User{ID_User: userCounter, Username: TempUser.Username, Password_Hash: passwordHash, Email: TempUser.Email, Is_Admin: true}

	//Теперь всё это передаём в бд
	connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
	db, _ := sql.Open("postgres", connStr)
	defer db.Close()

	sqlStatement := `INSERT INTO users (USERNAME, PASSWORD, EMAIL)
	VALUES ($1, $2, $3);`
	_, sqlErr := db.Exec(sqlStatement, TempUser.Username, passwordHash, TempUser.Email)
	if sqlErr != nil {
		fmt.Println(sqlErr)
	}
	w.Write([]byte(token)) //Сразу кидаем токен клиенту
	fmt.Println(usersMap)
	nc.Publish("log", []byte("command: registration, userId: "+strconv.Itoa(userCounter)+" username: "+TempUser.Username))
}

func addMonitorHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Connecting...")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	token := r.Header.Get("Authorization")
	if usersMap[token].Is_Admin {
		Monitor := MonitorTable{}
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()

		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &Monitor)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		sqlStatement := `INSERT INTO monitors (DISPLAY_ID, GSYNC_PREMIUM, CURVED)
	VALUES ($1, $2, $3);`
		_, sqlErr := db.Exec(sqlStatement, Monitor.DisplayID, Monitor.GSyncPremium, Monitor.IsCurved)
		if sqlErr != nil {
			fmt.Println(sqlErr)
		}
		nc.Publish("log", []byte("command: AddMonitor, userId: "+strconv.Itoa(usersMap[token].ID_User)+" Display_ID: "+strconv.Itoa(Monitor.DisplayID)+" Gsync_Premium: "+strconv.FormatBool(Monitor.GSyncPremium)+" Curved: "+strconv.FormatBool(Monitor.IsCurved)+"Отправил Зарипов"))
	} else {
		nc.Publish("log", []byte("Error. Admin not found. Token: "+token))
		w.WriteHeader(403)
	}
}

func addDisplayHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Connecting...")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()

	token := r.Header.Get("Authorization")
	if usersMap[token].Is_Admin {
		Display := DisplayTable{} // Складываем сюда дисплей из запроса
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()

		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &Display)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		sqlStatement := `INSERT INTO displays (DIAGONAL, RESOLUTION, TYPE, GSYNC)
	VALUES ($1, $2, $3, $4);`

		_, sqlErr := db.Exec(sqlStatement, Display.Diagonal, Display.Resolution, Display.Type, Display.GSync)
		if sqlErr != nil {
			fmt.Println(sqlErr)
		}
		nc.Publish("log", []byte("command: addDisplay, userId: "+strconv.Itoa(usersMap[token].ID_User)+" Diagonal: "+strconv.Itoa(int(Display.Diagonal))+" Resolution: "+Display.Resolution+" Type: "+Display.Type+" Gsync "+strconv.FormatBool(Display.GSync)))
	} else {
		nc.Publish("log", []byte("Error. Admin not found. Token: "+token))
		w.WriteHeader(403)
	}
}
