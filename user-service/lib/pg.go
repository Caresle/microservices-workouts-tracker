package lib

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func Pg(query string, params ...any) pgx.Rows {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	}

	defer conn.Close(context.Background())

	var result pgx.Rows

	if len(params) > 0 {
		result, err = conn.Query(context.Background(), query, params...)
	} else {
		result, err = conn.Query(context.Background(), query)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to query database: %v\n", err)
	}

	return result
}
