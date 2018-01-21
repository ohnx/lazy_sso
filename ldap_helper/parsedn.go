package ldap_helper

import (
    "strings"
)

type (
    DN struct {
        Cn      string
        Uid     string
        Ou      string
        Dc      []string
    }
)

func ParseDN(rawDN string) DN {
    var myDN DN

    // Get array of form "key1=value1", "key2=value2"
    kvs := strings.Split(rawDN, ",")

    myDN.Cn = ""

    for _, kv := range kvs {
        // Get array of form "key", "value"
        kvp := strings.Split(kv, "=")

        switch kvp[0] {
        case "uid":
            myDN.Uid = kvp[1]
        case "ou":
            myDN.Ou = kvp[1]
        case "dc":
            myDN.Dc = append(myDN.Dc, kvp[1])
        }
    }

    return myDN
}

func (q DN) String() string {
    str := "uid="+q.Uid+",ou="+q.Ou+",dc="+strings.Join(q.Dc,",dc=")
    if len(q.Cn) == 0 {
        return str
    } else {
        return "cn="+q.Cn+","+str
    }
}
