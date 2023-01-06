package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)


var Conn *pgx.Conn

func DataBaseConnection() {
	var err error
	databaseUrl := "postgres://postgres:AtlantaBig1738@localhost:5432/personal_web_adit"
	Conn, err = pgx.Connect(context.Background(),databaseUrl)
	if err != nil {
		
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v",err)
		os.Exit(1)
	}
	fmt.Println("Database connected")
}

