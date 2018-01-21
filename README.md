# lazy_sso

Lazy Single-Sign On (SSO) using LDAP as a protocol and a simple SQLite server
as a data store.

## Gogs configuration
```
User DN: uid=%s,ou=users,dc=changeme
User Filter: (&(objectClass=person)(uid=%s))
Admin Filter: (&(admin=yes)(uid=%s))
First Name Attribute: displayName
Email Attribute: mail
```

## Jenkins configuration


## More??

Please let me know and I'll look into how to add more :)
