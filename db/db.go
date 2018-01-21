package db

import (
    // standard library
    "log"
    "os"

    // hashing
    "crypto/sha512"
    "encoding/hex"

    // Database stuff
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type (
    User struct {
        Uid     string
        Admin   bool
        Mail    string
        Cn      string
    }
)

var initStmt = `
CREATE TABLE users (
    uid varchar,
    password varchar,
    admin boolean,
    mail varchar,
    cn varchar
);
`

var (
    // connection handle
    conn *sql.DB
)

func file_exists(filename string) bool {
    _, err := os.Stat(filename)

    if os.IsNotExist(err) {
        return false
    }

    return err == nil
}

func connect(filename string) {
    var err error
    conn, err = sql.Open("sqlite3", filename)
    if err != nil {
        log.Fatalf("Failed to open database: %s", err)
    }
}

func InitializeDatabase(filename string) {
    connect(filename)
    defer Disconnect()
    var err error

    log.Println("Info: Database file not found, creating new database...")

    // Initialize all the tables
    _, err = conn.Exec(initStmt)
    if err != nil {
        log.Fatalf("Failed to initialize database: %s", err)
    }

    // Add admin user
    stmt, err := conn.Prepare("INSERT INTO users(uid, password, admin, mail, cn) values(?,?,?,?,?)")
    if err != nil {
        log.Fatalf("Failed to initialize database: %s", err)
    }

    // TODO: configurable
    _, err = stmt.Exec("ohnx", Hash("password"), true, "me@masonx.ca", "Mason X")
    if err != nil {
        log.Fatalf("Failed to initialize database: %s", err)
    }

    // Be responsible!
    stmt.Close()
}

func Hash(str string) string {
    // SHA512 is the password hash being used
    hashalgo := sha512.New()
    hashalgo.Write([]byte(str))
    return hex.EncodeToString(hashalgo.Sum(nil))
}

func UserInDB(username string, password string) bool {
    // prepare read statement
    stmt, err := conn.Prepare("SELECT uid FROM users WHERE uid = ? AND password = ?")
    if err != nil {
        log.Printf("Warning: Failed to read database: %s", err)
        return false
    }
    defer stmt.Close()

    // Execute read statement
    res, err := stmt.Query(username, Hash(password))
    if err != nil {
        log.Printf("Warning: Failed to read database: %s", err)
        return false
    }
    defer res.Close()

    // Check results
    for res.Next() {
        var uid string
        // Only care about the 1st result
        err = res.Scan(&uid)
        if err != nil {
            log.Printf("Warning: Failed to read database: %s", err)
            return false
        }
        return true
    }
    // Incorrect username or password
    return false
}

func FetchUser(username string) (User, bool) {
    var myUser User
    // prepare read statement
    stmt, err := conn.Prepare("SELECT uid, admin, mail, cn FROM users WHERE uid = ?")
    if err != nil {
        log.Printf("Warning: Failed to read database: %s", err)
        return myUser, true
    }
    defer stmt.Close()

    // Execute read statement
    res, err := stmt.Query(username)
    if err != nil {
        log.Printf("Warning: Failed to read database: %s", err)
        return myUser, true
    }
    defer res.Close()

    // Check results
    for res.Next() {
        // Only care about the 1st result
        err = res.Scan(&myUser.Uid, &myUser.Admin, &myUser.Mail, &myUser.Cn)
        if err != nil {
            log.Printf("Warning: Failed to read database: %s", err)
            return myUser, true
        }

        return myUser, false
    }
    // Incorrect username or password
    return myUser, true
}

func Connect(filename string) {
    // Fill the database with blank tables if it doesn't exist
    if !file_exists(filename) {
        InitializeDatabase(filename)
    }

    connect(filename)
}

func Disconnect() {
    conn.Close()
}
