package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"../myutils"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

// main function to boot up everything
func main() {
	router := mux.NewRouter()
	log.Println("Starting listner on port 8080")

	router.HandleFunc("/store", getStore).Methods("GET")
	router.HandleFunc("/storeName", getStoreByName).Methods("GET")
	router.HandleFunc("/storesAll", getAllStores).Methods("GET")
	router.HandleFunc("/addStore", addStore).Methods("POST")
	router.HandleFunc("/addMockStores", addMockStores).Methods("GET")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	h1 := handlers.CombinedLoggingHandler(os.Stdout, router)
	h2 := handlers.CompressHandler(h1)
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(h2)))
	log.Println("Started listner on port 8080")
}

func getStore(w http.ResponseWriter, r *http.Request) {
	result := []myutils.Store{}
	code := r.URL.Query().Get("code")
	site := r.URL.Query().Get("site")

	if site == "" {
		log.Println("API site is a mandatory attribute")
		json.NewEncoder(w).Encode("API site is a mandatory attribute")
		return
	}

	if code == "" {
		log.Println("Store code is a mandatory attribute")
		json.NewEncoder(w).Encode("store code is a mandatory attribute")
		return
	}

	session := myutils.GetMongoDBSession(site)
	defer session.Close()

	c := session.DB(myutils.MongoDbDatabase[site]).C(myutils.MongoDbStoresCol)

	query := bson.M{"code": code}

	err := c.Find(query).All(&result)
	myutils.CheckErr(err)

	if len(result) > 0 {
		json.NewEncoder(w).Encode(result[0])
	} else {
		json.NewEncoder(w).Encode("no store found")
	}
}

func getStoreByName(w http.ResponseWriter, r *http.Request) {
	result := []myutils.Store{}
	name := r.URL.Query().Get("name")
	site := r.URL.Query().Get("site")

	if site == "" {
		log.Println("API site is a mandatory attribute")
		json.NewEncoder(w).Encode("API site is a mandatory attribute")
		return
	}

	if name == "" {
		log.Println("Store name is a mandatory attribute")
		json.NewEncoder(w).Encode("store name is a mandatory attribute")
		return
	}

	session := myutils.GetMongoDBSession(site)
	defer session.Close()

	c := session.DB(myutils.MongoDbDatabase[site]).C(myutils.MongoDbStoresCol)

	query := bson.M{"name": name}

	err := c.Find(query).All(&result)
	myutils.CheckErr(err)

	if len(result) > 0 {
		json.NewEncoder(w).Encode(result[0])
	} else {
		json.NewEncoder(w).Encode("no store found")
	}
}

func getAllStores(w http.ResponseWriter, r *http.Request) {
	site := r.URL.Query().Get("site")

	if site == "" {
		log.Println("API site is a mandatory attribute")
		json.NewEncoder(w).Encode("API site is a mandatory attribute")
		return
	}

	result := []myutils.Store{}
	session := myutils.GetMongoDBSession(site)
	defer session.Close()

	c := session.DB(myutils.MongoDbDatabase[site]).C(myutils.MongoDbStoresCol)

	err := c.Find(bson.M{}).All(&result)
	myutils.LogErr(err)

	json.NewEncoder(w).Encode(result)
}

func addStore(w http.ResponseWriter, r *http.Request) {
	site := r.URL.Query().Get("site")

	if site == "" {
		log.Println("API site is a mandatory attribute")
		json.NewEncoder(w).Encode("API site is a mandatory attribute")
		return
	}
	// Parse request
	decoder := json.NewDecoder(r.Body)
	var store myutils.Store
	err := decoder.Decode(&store)
	myutils.CheckErr(err)

	log.Println(store)

	session := myutils.GetMongoDBSession(site)
	defer session.Close()

	c := session.DB(myutils.MongoDbDatabase[site]).C(myutils.MongoDbStoresCol)

	_, err = c.Upsert(bson.M{"code": store.Code}, store)
	myutils.LogErr(err)
}

func addMockStores(w http.ResponseWriter, r *http.Request) {
	site := r.URL.Query().Get("site")

	if site == "" {
		log.Println("API site is a mandatory attribute")
		json.NewEncoder(w).Encode("API site is a mandatory attribute")
		return
	}

	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		fmt.Println("Error getting 'count' variable")
		count = 10
	}

	log.Println("Value of 'count' is:", count)

	session := myutils.GetMongoDBSession(site)
	defer session.Close()

	c := session.DB(myutils.MongoDbDatabase[site]).C(myutils.MongoDbStoresCol)

	templateStore := myutils.Store{
		Code:        "store_code",
		Name:        "Store Name ",
		Description: "Store description ",
		Catalogs:    []string{"c"},
		Area:        "Area",
		Delete:      false,
	}

	for i := 0; i < count; i++ {

		store := myutils.Store{
			Code:        templateStore.Code + strconv.Itoa(i),
			Name:        templateStore.Name + strconv.Itoa(i),
			Description: templateStore.Description + strconv.Itoa(i),
			Catalogs:    []string{templateStore.Catalogs[0] + strconv.Itoa(i)},
			Area:        templateStore.Area + strconv.Itoa((i%20)+1),
			Delete:      false,
		}

		_, err := c.Upsert(bson.M{"code": store.Code}, store)
		myutils.LogErr(err)

		log.Println("Inserted store number", i)
	}
}
