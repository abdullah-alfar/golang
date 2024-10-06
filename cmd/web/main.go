package main

import (
	"database/sql"
	"final-project/data"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

func main() {

	// connect to DB
	db := initDB()

	// create session

	session := initSession()

	//create logger
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create wait group

	wg := sync.WaitGroup{}

	// set up application config

	app := Config{
		Session:  session,
		DB:       db,
		Wait:     &wg,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Models:   data.New(db),
	}

	//listen for signal

	go app.listenForShutdown()

	//listen for web server

	app.serve()

}

func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panic("cannot connect to database")
	}
	return conn
}

func connectToDB() *sql.DB {
	counts := 0
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=password dbname=concurrency sslmode=disable timezone=UTC connect_timeout=5"
	}

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgress not yet ready ...")
		} else {
			log.Println("connected to DB")
			return connection
		}

		if counts > 10 {
			return nil
		}

		log.Println("backing off for 1 secobd")
		time.Sleep(1 * time.Second)
		counts++
		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
	fmt.Println("-----------------------------------------------")
	fmt.Println("DSN:", dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil

}

func initSession() *scs.SessionManager {
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true
	return session
}

func initRedis() *redis.Pool {
	redisPoll := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPoll
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	app.InfoLog.Printf("Starting web server ...")
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.shutdown()
	app.InfoLog.Println("Shutting down")
	os.Exit(0)
}

func (app *Config) shutdown() {
	app.InfoLog.Println("would run cleanup tasks...")
	app.Wait.Wait()

	app.InfoLog.Println("closing")

}
