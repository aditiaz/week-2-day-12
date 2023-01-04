package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

// pointer * menseleksi semua yg di var tsb
var Conn *pgx.Conn

func DataBaseConnection() {
	var err error
	databaseUrl := "postgres://postgres:pHere@localhost:5432/personal_web_adit"
	// adac kosong konteksnya
	Conn, err = pgx.Connect(context.Background(),databaseUrl)
	if err != nil {
		// %v passing valuenya ke err
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v",err)
		os.Exit(1)
	}
	fmt.Println("Database connected")
}

// postgres://{user}:{password}@{host}:{port}/{database}
// host =postgressnya
// port= port postgress
// database = database postgress