package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"

    blah "github.com/vjeantet/goldap/message"
    ldap "github.com/vjeantet/ldapserver"
    "github.com/ohnx/lazy_sso/db"
    "github.com/ohnx/lazy_sso/ldap_helper"
)

func main() {
    var host string
    //var basedn string
    var dbfn string
    var password string

    flag.StringVar(&host, "host", "0.0.0.0:10389", "Host to listen on")
    // NYI - flag.StringVar(&basedn, "basedn", "example.com", "Base DN")
    flag.StringVar(&dbfn, "db", "data.db", "Database file to use")
    flag.StringVar(&password, "password", "", "Specify this to hash a password using the database hashing method ONLY. (program exits after)")

    if len(password) != 0 {
        fmt.Printf("Hashed password: ", db.Hash(password))
        return
    }

    // Connect to database
    db.Connect(dbfn)

    // Logger
    ldap.Logger = log.New(os.Stdout, "[server] ", log.LstdFlags)

    // Create a new LDAP Server
    server := ldap.NewServer()

    routes := ldap.NewRouteMux()
    routes.Bind(handleBind)
    routes.Search(handleSearch).Label("Search")
    server.Handle(routes)

    // listen on 10389
    go server.ListenAndServe(host)

    // When CTRL+C, SIGINT and SIGTERM signal occurs
    // Then stop server gracefully
    ch := make(chan os.Signal)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    <-ch
    close(ch)

    server.Stop()
}

// handleBind return Success if login == mysql
func handleBind(w ldap.ResponseWriter, m *ldap.Message) {
    r := m.GetBindRequest()
    res := ldap.NewBindResponse(ldap.LDAPResultSuccess)

    dn := ldap_helper.ParseDN(string(r.Name()))

    if (len(dn.Uid) == 0 && len(string(r.AuthenticationSimple())) == 0) ||
        db.UserInDB(dn.Uid, string(r.AuthenticationSimple())) {
        w.Write(res)
        return
    }

    ldap.Logger.Printf("Authentication failed for user %s", string(r.Name()))
    // debug
    log.Printf("Password: %s", string(r.AuthenticationSimple()))
    res.SetResultCode(ldap.LDAPResultInvalidCredentials)
    res.SetDiagnosticMessage("invalid credentials")
    w.Write(res)
}

func handleSearch(w ldap.ResponseWriter, m *ldap.Message) {
    r := m.GetSearchRequest()
    dn := ldap_helper.ParseDN(string(r.BaseObject()))

    // debug
    log.Printf("Request BaseDn=%s", r.BaseObject())
    log.Printf("Request Filter=%s", r.Filter())
    log.Printf("Request FilterString=%s", r.FilterString())
    log.Printf("Request Attributes=%s", r.Attributes())
    log.Printf("Request TimeLimit=%d", r.TimeLimit().Int())

    // Handle Stop Signal (server stop / client disconnected / Abandoned request....)
    select {
    case <-m.Done:
        return
    default:
    }

    // Get user
    user, haderr := db.FetchUser(dn.Uid)

    if haderr == true {
        // Failed to find the user
    } else if strings.Contains(string(r.FilterString()), "admin") {
        // Case 2: looking for user admin-ness
        ldap.Logger.Printf("Requesting info if %s is admin", dn.Uid)

        // Only spit back output if user is actually admin
        if user.Admin {
            // Return stuff
            dn.Cn = user.Cn;
            e := ldap.NewSearchResultEntry(dn.String())
            e.AddAttribute("mail", blah.AttributeValue(user.Mail))
            e.AddAttribute("cn", blah.AttributeValue(user.Cn))
            w.Write(e)
            fmt.Println(e)
        }
    } else {
        // Case 1: looking for user info
        ldap.Logger.Printf("Requesting info for %s", dn.Uid)

        // Return stuff
        dn.Cn = user.Cn;
        e := ldap.NewSearchResultEntry(dn.String())
        e.AddAttribute("mail", blah.AttributeValue(user.Mail))
        e.AddAttribute("cn", blah.AttributeValue(user.Cn))
        e.AddAttribute("admin", "yes")
        w.Write(e)
        fmt.Println(e)
    }

    res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
    w.Write(res)
}
