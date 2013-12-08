# go-todo
Go-Todo is a simple REST web application for managing your todo/task lists, written in Go.      
       
## Overview
Go-Todo contains 2 objects, accounts and tasks.       
An account can have many tasks.
Accounts can be set to 3 different roles:      
 - Admin  
 - User  
 - None  

## Installation
Make sure you have a working Go environment (*Requires* Go1.2+).   
([Offical installation instructions](http://golang.org/doc/install.html))

Use go get to install go-todo:
```
$ go get github.com/JamesClonk/go-todo
```

Make sure your PATH includes the `$GOPATH/bin` directory:
```
export PATH=$PATH:$GOPATH/bin
```

## Usage
Run go-todo: 

```
$ go-todo
```

It will now start a webserver listening on port 8008, and provide a REST interface with the following endpoints:  
 - /auth  
 - /tasks  
 - /task/{taskId}  
 - /accounts  
 - /account/{accountId}  

Use *GET* on **/auth** with query parameter ?login={email} to retrieve auth information for a particular user account.     
If provided a valid email will return the account id, the server timestamp and the account salt.      

All following requests need to be given the query string:       
?rId={accountId}     
&rTimestamp={current-timestamp-offset-by-server-timestamp}     
&rSalt={random-string}     
&rToken={sha512-hash-of(rTimestamp+rSalt+sha512-hash-of(accountSalt+accountPassword)))}     

*GET* on **/tasks** will return a list of all tasks belonging to the account used in the request.        
(Even an account with role "Admin" only gets his tasks returned)

*GET*, *POST*, *PUT* and *DELETE* on **/task/{taskId}** pretty much do what you'd expect.      
(The account your using needs to be either the owner of these tasks for GET, PUT and DELETE, or needs to have the "Admin" role)

*GET* on **/accounts** will return a list of all accounts in the db.      
(Only an "Admin" account can request this)

*GET*, *POST*, *PUT* and *DELETE* on **/account/{accountId}** also somewhat does what you'd expect.      
(Most things here only work or make sense using an account with "Admin" role)

## Client
Next step for me to do is to write a simple standalone html file using jQuery to act as demo client.     
It's on my Todo-List. ;-)

## License
GPL v3, see LICENSE.TXT      



