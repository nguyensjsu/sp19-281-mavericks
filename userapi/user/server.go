/**
* API for accessing and manipulating user information
 */

package main

/* Imports */
import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"github.com/unrolled/render"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	// "net"
	// "os"
	// "strings"
)

/*var mongodb_server = os.Getenv("MONGO_SERVER")
var mongodb_database = os.Getenv("MONGO_DATABASE")
var mongodb_collection = os.Getenv("MONGO_COLLECTION")
var mongo_admin_database = os.Getenv("MONGO_ADMIN_DATABASE")
var mongo_username = os.Getenv("MONGO_USERNAME")
var mongo_password = os.Getenv("MONGO_PASS")*/

var mongodb_server = "localhost"
var mongodb_database = "userdb"
var mongodb_collection = "users"

func newUserServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	n := negroni.Classic()
	router := mux.NewRouter()
	initRoutes(router, formatter)
	n.UseHandler(router)
	return n
}

/* Initializing resource URI */
func initRoutes(router *mux.Router, formatter *render.Render) {
	router.HandleFunc("/users", getAllUsers).Methods("GET")
	router.HandleFunc("/users/{id}", getUser).Methods("GET")
	// router.HandleFunc("/users/signup", createUser).Methods("POST")
	// router.HandleFunc("/users/signin", userSignIn).Methods("POST")
	// router.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	// router.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/ping", checkPing(formatter)).Methods("GET")
}

/* Setup response headers */
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

/* Handler for /ping */
func checkPing(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		message := "User API is alive!"
		formatter.JSON(w, http.StatusOK, struct{ Test string }{message})
	}
}

/* Handler for /user to fetch all users */
func getAllUsers(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	w.Header().Set("Content-Type", "application/json")

	/* Open DB session */
	session, err := mgo.Dial(mongodb_server)
	if err != nil {
		message := struct{ Message string }{"Error while connecting to database"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	var result []User // Result set to store records

	/* Open DB connection*/
	c := session.DB(mongodb_database).C(mongodb_collection)

	query := bson.M{} // Empty query to fetch all records

	err = c.Find(query).All(&result) // Fetch all users

	if err != nil {
		message := struct{ Message string }{"Error while fetching users"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}

	json.NewEncoder(w).Encode(result)
}

/**
 * Fetch record of given user
 * Handler for /user/{id}
 */
func getUser(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	w.Header().Set("Content-Type", "application/json")

	/* Open DB session */
	session, err := mgo.Dial(mongodb_server)
	if err != nil {
		message := struct{ Message string }{"Error while connecting to database"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	params := mux.Vars(req)

	var id string = params["id"]

	query := bson.M{"id": id}

	var result bson.M

	if id == "" {
		message := struct{ Message string }{"User ID not provided"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(message)
		return
	} else {
		/* Obtain DB connection */
		c := session.DB(mongodb_database).C(mongodb_collection)

		/* Execute query */
		err = c.Find(query).One(&result)

		/* Handle errors and return apt response*/
		if err != nil && err != mgo.ErrNotFound {
			text := "Error while fetching user information" + id
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(text)
			return
		}

		if err != nil && err == mgo.ErrNotFound {
			text := "No user found for ID: " + id
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(text)
			return
		}
	}

	json.NewEncoder(w).Encode(result)
}

func createUser(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	w.Header().Set("Content-Type", "application/json")

	var user User
	_ = json.NewDecoder(req.Body).Decode(&user)
	uniqueId := uuid.NewV4()
	user.Id = uniqueId.String()

	/* Open DB session */
	session, err := mgo.Dial(mongodb_server)
	if err != nil {
		message := "Error while connecting to database"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	/* Obtain DB connection */
	c := session.DB(mongodb_database).C(mongodb_collection)

	/* TODO Validate that email ID is not empty */

	/* Check for duplicate email address */
	query := bson.M{"email": user.Email}
	var result bson.M
	err = c.Find(query).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		message := "Error while fetching data"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	} else if result != nil {
		message := "User with this email ID already exists"
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(message)
		return
	}

	/* Commit new user info */
	err = c.Insert(user)
	if err != nil {
		message := "Error while creating new user"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}

	/* Return newly created user */
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

/*
func deleteUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	session, err := mgo.Dial(mongodb_server)
	if err != nil {
		message := struct{ Message string }{"Some error occured while connecting to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}
	err = session.DB(mongo_admin_database).Login(mongo_username, mongo_password)
	if err != nil {
		message := struct{ Message string }{"Some error occured while login to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mongodb_database).C(mongodb_collection)
	query := bson.M{"id": params["id"]}
	err = c.Remove(query)
	if err != nil && err != mgo.ErrNotFound {
		message := struct{ Message string }{"Some error occured while querying to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		fmt.Println("error:" + err.Error())
		return
	} else if err == mgo.ErrNotFound {
		message := struct{ Message string }{"user not found!!"}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(message)
		return
	}
	json.NewEncoder(w).Encode(struct{ Message string }{"user with id:" + params["id"] + " was deleted"})
}
func updateUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var person User
	_ = json.NewDecoder(req.Body).Decode(&person)
	params := mux.Vars(req)
	session, err := mgo.Dial(mongodb_server)
	if err != nil {
		message := struct{ Message string }{"Some error occured while connecting to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}
	err = session.DB(mongo_admin_database).Login(mongo_username, mongo_password)
	if err != nil {
		message := struct{ Message string }{"Some error occured while login to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mongodb_database).C(mongodb_collection)
	query := bson.M{"id": params["id"]}
	updator := bson.M{
		"$set": bson.M{
			"firstname": person.Firstname,
			"lastname":  person.Lastname,
			"address":   person.Address,
			"password":  person.Password}}
	err = c.Update(query, updator)
	if err != nil && err != mgo.ErrNotFound {
		message := struct{ Message string }{"Some error occured while querying to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	} else if err == mgo.ErrNotFound {
		message := struct{ Message string }{"User not found!!"}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(message)
		return
	}
	json.NewEncoder(w).Encode(struct{ Message string }{"user with id:" + params["id"] + " was Updated"})
}
func userSignIn(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var person User
	_ = json.NewDecoder(req.Body).Decode(&person)
	session, err := mgo.Dial(mongodb_server)
	if err != nil {
		message := struct{ Message string }{"Some error occured while connecting to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}
	err = session.DB(mongo_admin_database).Login(mongo_username, mongo_password)
	if err != nil {
		message := struct{ Message string }{"Some error occured while login to database!!"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(message)
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mongodb_database).C(mongodb_collection)
	query := bson.M{"email": person.Email,
		"password": person.Password}
	var result User
	err = c.Find(query).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		message := struct{ Message string }{"Some error occured while querying to database!!"}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(message)
		return
	}
	if err == mgo.ErrNotFound {
		message := struct{ Message string }{"Login Failed"}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(message)
		return
	}
	userData := bson.M{
		"email":     result.Email,
		"firstName": result.Firstname,
		"lastName":  result.Lastname,
		"address":   result.Address,
		"id":        result.Id}
	json.NewEncoder(w).Encode(userData)
}*/