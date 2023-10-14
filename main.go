package main

import (
	"highload-arch/pkg/backend"
	"highload-arch/pkg/config"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
)

/*func RunReq() error {
	db, err := pgx.Connect(context.Background(), "host=localhost port=5432 user=admin_user password=1111 dbname=social_net sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	regexMap := make(map[string]string)
	regexMap["first_name"] = string('\'') + string(rune('У')) + string('%') + string('\'')
	regexMap["second_name"] = string('\'') + string(rune('Р')) + string('%') + string('\'')
	regexFilter := ``
	var regexMapArray []interface{}
	id := 1
	for key, value := range regexMap {
		newFilter := fmt.Sprintf(`%s LIKE %s`, key, value)
		if regexFilter != `` {
			regexFilter += ` and ` + newFilter
		} else {
			regexFilter = newFilter
			id += 1
		}
		regexMapArray = append(regexMapArray, value)
	}
	regexFilter = `SELECT * FROM users WHERE ` + regexFilter + ` ORDER BY id`
	fmt.Println(regexFilter)
	start := time.Now()
	rows, err := db.Query(context.Background(), regexFilter)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
		return err
	}
	elapsed := time.Since(start)
	fmt.Printf("Execution of Select: %v\n", elapsed)
	return nil
}*/

func main() {
	config.Load("local-config.yaml")
	storage.CreateConnectionPool()

	log.Printf("Server started")
	router := backend.NewRouter()
	log.Fatal(http.ListenAndServe(config.GetString("server.port"), router))

	//server := echo.New()
	//server.HideBanner = true

	// Routes
	//backend.AddRoutes(server)

	// Start REST server
	//err := server.Start(config.GetString("server.port"))
	//server.Logger.Fatal(err)

}
