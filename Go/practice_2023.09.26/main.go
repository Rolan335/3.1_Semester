package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/lib/pq"
)

var usersMap = make(map[string]User)
var userCounter = 0

type DisplayInfo struct {
	Diagonal   float32 `json:"diagonal"`
	Resolution string  `json:"resolution"`
	Type       string  `json:"type"`
	GSync      bool    `json:"gsync"`
}

type MonitorInfo struct {
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

func main() {
	// При запуске сервера, сгружаем все логины из бд в мапу, чтобы в дальнейшем к ним обращаться
	connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
	db, _ := sql.Open("postgres", connStr)
	defer db.Close()
	rows, _ := db.Query("SELECT id_user,username, password, email, is_admin FROM users;")
	for rows.Next() {
		userCounter++
		var rowSlice User
		err := rows.Scan(&rowSlice.ID_User, &rowSlice.Username, &rowSlice.Password_Hash, &rowSlice.Email, &rowSlice.Is_Admin)
		if err != nil {
			fmt.Println(err)
		}
		//генерируем токен и добавляем ключ:значение в мапу
		// токен = string(sha256(username + password_hash))
		token := fmt.Sprintf("%x", sha256.Sum256([]byte(rowSlice.Username+rowSlice.Password_Hash)))
		usersMap[token] = rowSlice
	}
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

	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		panic(err)
	}
}

func getDisplayById(w http.ResponseWriter, r *http.Request) {
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
		var displays []DisplayInfo
		for rows.Next() {
			var rowSlice DisplayInfo
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

	} else {
		w.WriteHeader(403)
	}
}

func deleteDisplayHandler(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(200)
	} else {
		w.WriteHeader(403)
	}
}

func deleteMonitorHandler(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(200)
	} else {
		w.WriteHeader(403)
	}
}

// TODO : Сделать ручку добавления админа
func makeAdmin(w http.ResponseWriter, r *http.Request) {
	// token := r.Header.Get("Authorization")
}

func allMonitorsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if _, ok := usersMap[token]; ok {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()
		rows, _ := db.Query("SELECT display_id, gsync_premium, curved FROM monitors;")
		var monitors []MonitorInfo
		for rows.Next() {
			var rowSlice MonitorInfo
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
	} else {
		w.WriteHeader(403)
	}
}

func allDisplaysHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if _, ok := usersMap[token]; ok {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		connStr := "postgresql://postgres:Pa$$w0rd@localhost:5432/postgres?sslmode=disable"
		db, _ := sql.Open("postgres", connStr)
		defer db.Close()
		rows, _ := db.Query("SELECT diagonal, resolution, type, gsync FROM displays;")
		var displays []DisplayInfo
		for rows.Next() {
			var rowSlice DisplayInfo
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
	} else {
		w.WriteHeader(403)
	}
}

// Проверяем, зарегистрирован ли пользователь по нашей мапе
// Если да, отправляем клиенту его токен.
func login(w http.ResponseWriter, r *http.Request) {
	TempUser := UserLogin{}
	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &TempUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Password)))
	token := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Username+passwordHash)))
	stringToken := fmt.Sprintf("%x", token)
	_, ok := usersMap[stringToken]
	if ok {
		w.Write([]byte(stringToken))
	} else {
		w.WriteHeader(403)
	}
}

func registration(w http.ResponseWriter, r *http.Request) {
	TempUser := UserRegistration{}
	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &TempUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Password))) //Хэшированный пароль

	//Генерируем токен и кидаем в мапу
	userCounter += 1
	token := fmt.Sprintf("%x", sha256.Sum256([]byte(TempUser.Username+passwordHash)))
	usersMap[token] = User{ID_User: userCounter, Username: TempUser.Username, Password_Hash: passwordHash, Email: TempUser.Email, Is_Admin: false}

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
}

func addMonitorHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if usersMap[token].Is_Admin {
		Monitor := MonitorInfo{}
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
	} else {
		w.WriteHeader(403)
	}
}

func addDisplayHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if usersMap[token].Is_Admin {
		Display := DisplayInfo{} // Складываем сюда дисплей из запроса
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
	} else {
		w.WriteHeader(403)
	}
}
