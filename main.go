package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"strconv"
)

var coffeeDB *sql.DB

type Coffee struct {
	ID      int64    `json:"id"`
	Roaster string   `json:"roaster"`
	RoasterLocation sql.NullString   `json:"roaster_location"`
	Name    string   `json:"name"`
	Origins sql.NullString `json:"origins"`
	ImageURL sql.NullString `json:"imageurl"`
}

type Machine struct {
	ID      int64    `json:"id"`
	Vendor string   `json:"vendor"`
	Name    string   `json:"name"`
	ImageURL sql.NullString `json:"imageurl"`
}


func main() {
	db, err := sql.Open("sqlite3", "coffees.db")

	if err != nil {
		log.Fatal("Could not open DB file: ", err)
	}
	coffeeDB = db

	rows, err := coffeeDB.Query("SELECT * FROM coffees")

	if (err != nil || rows == nil){
		_, err = db.Exec("CREATE TABLE `coffees` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"`name` VARCHAR(255) NOT NULL," +
			"`roaster` VARCHAR(255) NOT NULL, " +
			"`roaster_location` TEXT," +
			"`origin` TEXT," +
			"`image_url` TEXT)")
		if (err != nil){
			log.Fatal("Could not init DB file: ", err)
		}
	}

	rows, err = coffeeDB.Query("SELECT * FROM machines")

	if (err != nil || rows == nil){
		_, err = db.Exec("CREATE TABLE `machines` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"`vendor` TEXT," +
			"`name` TEXT," +
			"`image_url` TEXT)")
		if (err != nil){
			log.Fatal("Could not init DB file: ", err)

		}
	}
	router := gin.Default()
	api := router.Group("/api")
	v1 := api.Group("/v1")
	coffees := v1.Group("/coffees")
	coffees.GET("/:id", getCoffeeByID)
	coffees.PUT("/:id", updatCoffeeeByID)
	coffees.GET("", getAllCoffees)
	coffees.POST("", insertCoffee)

	roasters := v1.Group("/roasters")
	roasters.GET("", getAllRoasters)
	roasters.GET("/:roaster", getAllCoffeesByRoaster)

	machines := v1.Group("/machines")
	//machines.GET("/:id", getVendorByID)
	//machines.PUT("/:id", updatVendorByID)
	machines.GET("", getAllMachines)
	machines.POST("", insertMachine)

	http.Handle("/api/v1/", router)

	log.Print(" Running on 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	db.Close();
}

func getAllMachines(c *gin.Context) {
	rows, err := coffeeDB.Query("SELECT * FROM machines")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	var roasters []Machine
	for rows.Next() {
		var roaster Machine
		err = rows.Scan(&roaster.ID, &roaster.Vendor, &roaster.Name, &roaster.ImageURL)

		if err != nil {
			log.Print("Error during SQL scanning: ",err)
			c.IndentedJSON(http.StatusInternalServerError, nil)
			return
		}
		roasters = append(roasters, roaster)
	}

	c.IndentedJSON(http.StatusOK, roasters)
}

func insertMachine(c *gin.Context) {
	var machine Machine
	err := c.BindJSON(&machine)

	if err != nil {
		log.Print("Could not bind JSON! :",err)
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	enc:= json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(machine);

	stmt, err := coffeeDB.Prepare("SELECT * FROM machines where NAME=? AND VENDOR=?")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	rows, err := stmt.Query(machine.Name, machine.Vendor)
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}

	if rows.Next() {
		err = rows.Scan(&machine.ID, &machine.Vendor, &machine.Name, &machine.ImageURL)

		if err != nil {
			log.Print("Error during SQL scanning: ",err)
			c.IndentedJSON(http.StatusInternalServerError, nil)
			return
		}
		c.IndentedJSON(http.StatusOK, machine)
		return

	}

	stmt, err = coffeeDB.Prepare("INSERT INTO machines(vendor, name, image_url) values (?, ?, ?)")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}

	result, err := stmt.Exec(machine.Vendor, machine.Name, machine.ImageURL)
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	newID, err := result.LastInsertId()
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	machine.ID = newID

	c.IndentedJSON(http.StatusOK, machine)
}


func getAllRoasters(c *gin.Context) {
	rows, err := coffeeDB.Query("SELECT `roaster` FROM coffees")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	var roasters []string
	for rows.Next() {
		var roaster sql.NullString
		err = rows.Scan(&roaster)

		if err != nil {
			log.Print("Error during SQL scanning: ",err)
			c.IndentedJSON(http.StatusInternalServerError, nil)
			return
		}
		if roaster.Valid {
			roasters = append(roasters, roaster.String)
		}
	}

	c.IndentedJSON(http.StatusOK, roasters)
}

func getAllCoffees(c *gin.Context) {
	rows, err := coffeeDB.Query("SELECT * FROM coffees")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	var coffees []Coffee
	for rows.Next() {
		var coffee Coffee
		err = rows.Scan(&coffee.ID, &coffee.Name, &coffee.Roaster, &coffee.RoasterLocation, &coffee.Origins, &coffee.ImageURL)

		if err != nil {
			log.Print("Error during SQL scanning: ",err)
			c.IndentedJSON(http.StatusInternalServerError, nil)
			return
		}
		coffees = append(coffees, coffee)
	}

	c.IndentedJSON(http.StatusOK, coffees)
}

func getCoffeeByID(c *gin.Context) {
	id := c.Param("id")
	stmt, err := coffeeDB.Prepare(" SELECT * FROM coffees where id = ?")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	rows, err := stmt.Query(id)
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	var coffee Coffee
	for rows.Next() {
		err = rows.Scan(&coffee.ID, &coffee.Name, &coffee.Roaster, &coffee.RoasterLocation, &coffee.Origins, &coffee.ImageURL)
		if err != nil {
			log.Print("Error during SQL scanning: ",err)
			c.IndentedJSON(http.StatusInternalServerError, nil)
			return
		}
	}

	c.IndentedJSON(http.StatusOK, coffee)
}

func getAllCoffeesByRoaster(c *gin.Context) {
	roaster := c.Param("roaster")
	stmt, err := coffeeDB.Prepare(" SELECT * FROM coffees where roaster = ?")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	rows, err := stmt.Query(roaster)
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	var coffees []Coffee
	for rows.Next() {
		var coffee Coffee
		err = rows.Scan(&coffee.ID, &coffee.Name, &coffee.Roaster, &coffee.RoasterLocation, &coffee.Origins, &coffee.ImageURL)
		if err != nil {
			log.Print("Error during SQL scanning: ",err)
			c.IndentedJSON(http.StatusInternalServerError, nil)
			return
		}
		coffees = append(coffees, coffee)
	}

	c.IndentedJSON(http.StatusOK, coffees)
}


func insertCoffee(c *gin.Context) {
	var coffee Coffee
	c.BindJSON(&coffee)
	stmt, err := coffeeDB.Prepare("INSERT INTO coffees(roaster, name) values (?, ?)")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	result, err := stmt.Exec(coffee.Roaster, coffee.Name)
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	newID, err := result.LastInsertId()
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	coffee.ID = newID

	c.IndentedJSON(http.StatusOK, coffee)
}

func updatCoffeeeByID(c *gin.Context) {
	id := c.Param("id")
	var coffee Coffee
	ID, _ := strconv.ParseInt(id, 10, 0)
	coffee.ID = ID
	stmt, err := coffeeDB.Prepare("UPDATE coffees SET name = ? WHERE id = ?")
	if err != nil {
		log.Print("Error during SQL query: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	result, err := stmt.Exec(coffee.Name, coffee.ID)
	if err != nil {
		log.Print("Error during SQL exec: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	rowAffected, err := result.RowsAffected()
	if err != nil {
		log.Print("Error during SQL RowsAffected: ",err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	if rowAffected > 0 {
		c.IndentedJSON(http.StatusOK, coffee)
	} else {
		c.IndentedJSON(http.StatusNotFound, nil)
	}

}
