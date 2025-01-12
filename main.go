package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/usepzaka/fiberflash"
	"golang.org/x/crypto/bcrypt"
)

// The function to extract values from .env file
func viperEnv(key string) string {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Fatal error occured while reading the config file %s", err)
	}
	// Since our .env contains a key value pair values hence using key to get the value.
	value, ok := viper.Get(key).(string)
	if !ok {
		log.Fatalf("Invalid type assertion")
	}

	return value
}

func main() {
	// Postgres connection requirements
	host := "localhost"
	port := 5432

	// Database credenctials
	user := viperEnv("USERNAME")
	password := viperEnv("PASSWORD")
	dbname := viperEnv("DBNAME")

	// PostgreSQL connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Connecting to Postgres
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// Printing success message if everything went smooth.
	fmt.Println("Successfully connected!")

	// Creating an html template parser for fiber
	engine := html.New("./views", ".html")

	// Instanciating fiber with configurations
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Parser for static files like images which resides in public directory
	app.Static("/static", "./public")

	// The home route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":   "Welcome to Fiber!",
			"Message": "This is an example of rendering templates in Fiber.",
		})
	})

	// The login route
	app.Get("/login", func(c *fiber.Ctx) error {

		return c.Render("login", fiber.Map{
			"error":   false,
			"message": "Aok!",
		})
	})

	// A different route to handle fiberflash messages.
	app.Get("/error", func(c *fiber.Ctx) error {
		flash := fiberflash.Get(c)

		return c.Render("login", fiber.Map{
			"Flash": flash,
		})
	})

	app.Get("/success", func(c *fiber.Ctx) error {
		flash := fiberflash.Get(c)

		return c.Render("login", fiber.Map{
			"Flash": flash,
		})
	})

	// Register route.
	app.Get("/register/", func(c *fiber.Ctx) error {
		return c.Render("register", fiber.Map{
			"Title":   "Register me",
			"Message": "Register yourself here",
		})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		// User's Login credentials
		username := c.FormValue("username")
		password := c.FormValue("password")

		// Performing sql query to database to fetch the user details if exists
		sqlStatement := `
					SELECT name, password FROM users
					WHERE email = $1`

		// Creating variables to store the output returned from db.QueryRow
		var dbUsername, dbPassword string
		err = db.QueryRow(sqlStatement, username).Scan(&dbUsername, &dbPassword)
		if err != nil {
			fmt.Println("Error:", err)

			mp := fiber.Map{
				"error":   true,
				"message": "The email don't exists Please enter a valid email!",
			}

			return fiberflash.WithError(c, mp).Redirect("/error")

		}

		// Using bcrypt module to compare the hashed password string with the user typed normal password string
		err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
		if err != nil {
			mp := fiber.Map{
				"error":   true,
				"message": "The password didn't matched! Please enter a valid password",
			}
			return fiberflash.WithError(c, mp).Redirect("/error")
		}

		return c.Render("success", fiber.Map{
			"Title":   "Success",
			"Message": "Congratulations! You've logged in successfully",
		})
	})

	app.Post("/register_me", func(c *fiber.Ctx) error {
		// Fetching the form data entered by the user
		name := c.FormValue("name")
		email := c.FormValue("email")
		password := c.FormValue("password")

		// Converting the normal password string into byte form to be readable by bcrypt module.
		bytePassword := []byte(password)
		// Converting the byte password to hashed password.
		hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)

		if err != nil {
			panic(err)
		}

		// Inserting the user details into our database
		sqlStatement := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)`
		_, err = db.Exec(sqlStatement, name, email, string(hashedPassword))
		if err != nil {
			fmt.Println("Error executing query:", err)
			return c.Render("login", fiber.Map{
				"Title":   "failure",
				"Message": "Failed to log in",
			})
		}

		// Redirecting user to success page.
		mp := fiber.Map{
			"error":   true,
			"message": "Registered Successfully. Please Login!",
		}
		return fiberflash.WithError(c, mp).Redirect("/success")

	})

	app.Listen(":3000")

}
