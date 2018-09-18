package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"

	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sqlx.DB

type user struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

var userType = graphql.NewObject(
	graphql.ObjectConfig {
		Name: "User",
		Fields: graphql.Fields {
			"id": &graphql.Field {
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig {
		Name: "Query",
		Fields: graphql.Fields {
			"user": &graphql.Field {
				Type: userType,
				Args: graphql.FieldConfigArgument {
					"id": &graphql.ArgumentConfig {
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idQuery, isOK := p.Args["id"].(string)
					if isOK {
						var u = user{}
						err := DB.Get(&u, "SELECT * FROM user WHERE id=$1", idQuery)
						if err != nil {
							return nil, nil
						}
						return &u, nil
					}

					return nil, nil
				},
			},
		},
	},
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig {
		Query: queryType,
	},
)

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params {
		Schema: schema,
		RequestString: query,
	})

	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}

	return result
}

func handler(w http.ResponseWriter, r *http.Request) {
	bufBody := new(bytes.Buffer)
	bufBody.ReadFrom(r.Body)
	query := bufBody.String()

	result := executeQuery(query, schema)
	json.NewEncoder(w).Encode(result)
}

func main() {
	connectDB()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func connectDB() {
	db, err := sqlx.Connect("mysql", "root:@/sample_graphql")
	if err != nil {
		fmt.Print("Error:", err)
	}
	DB = db
}
