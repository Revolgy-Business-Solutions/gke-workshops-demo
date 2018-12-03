package main

import (
  "encoding/json"
  "log"
  "net/http"
  "github.com/gorilla/mux"
  "runtime"
  "time"
  "fmt"
  "os"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

type Book struct{
  ID      string  `json:"id,omitempty"`
  Name    string  `json:"name,omitempty"`
  Author  string  `json:"author,omitempty"`
  Pages   int     `json:"pagesomitempty"`
}

var books []Book
var SQLDBENABLED = os.Getenv("MYSQL_DATABASE")
var db *sql.DB

func main() {

  var err error
  if SQLDBENABLED == "True" {
    fmt.Println("sqldb")
    dbuser := os.Getenv("MYSQL_USER")
    dbpass := os.Getenv("MYSQL_PASS")

    db, err = sql.Open("mysql", dbuser + ":" + dbpass + "@tcp(127.0.0.1:3306)/bookstore")
    if err != nil {
      panic(err)
    }

  } else {
    fmt.Println("defaultdb")
    books = append(books, Book{ID: "1", Name: "Kubernetes up and running", Author: "Kelsey Hightower", Pages: 192})
    books = append(books, Book{ID: "2", Name: "Site Reliability Engineering", Author: "Betsy Beyer", Pages: 552})
    books = append(books, Book{ID: "3", Name: "Docker: Up & Running", Author: "Sean P Kane", Pages: 347})
  }

  router := mux.NewRouter()
  router.HandleFunc("/", GetRoot).Methods("GET")
  router.HandleFunc("/cpu", GetCpu).Methods("GET")
  router.HandleFunc("/kube", GetKube).Methods("GET")
  router.HandleFunc("/books", GetAllBooks).Methods("GET")
  router.HandleFunc("/books/{id}", GetBook).Methods("GET")
  router.HandleFunc("/books/{id}", CreateBook).Methods("POST")
  router.HandleFunc("/books/{id}", DeleteBook).Methods("DELETE")
  log.Fatal(http.ListenAndServe(":8000", router))

}

func GetRoot(w http.ResponseWriter, r *http.Request) {
}

func GetCpu(w http.ResponseWriter, r *http.Request) {
  done := make(chan int)

  for i := 0; i < runtime.NumCPU(); i++ {
    go func() {
      for {
          select {
          case <-done:
              return
          default:
          }
      }
    }()
  }

  time.Sleep(time.Second * 1)
  close(done)

  log.Println("Yay, I stressed!")
}

func GetKube(w http.ResponseWriter, r *http.Request) {

  http.ServeFile(w, r, "assets/kube.html")
}

func GetAllBooks(w http.ResponseWriter, r *http.Request) {
  if SQLDBENABLED == "True" {
    results, err := db.Query("SELECT * FROM books")
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
    }
    var books_res []Book

	  for results.Next() {
      var book Book
      // for each row, scan the result into our tag composite object
      err = results.Scan(&book.ID, &book.Name, &book.Author, &book.Pages)
      if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
      }
                  // and then print out the tag's Name attribute
     books_res = append(books_res, book)
	 }
   json.NewEncoder(w).Encode(books_res)

  } else {
     json.NewEncoder(w).Encode(books)
  }
}

func GetBook(w http.ResponseWriter, r *http.Request) {
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
}

