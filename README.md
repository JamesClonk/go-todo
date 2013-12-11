# go-todo
Go-Todo is a simple REST web application for managing your todo/task lists, written in Go.      
       
## Overview
Go-Todo contains 2 objects, accounts and tasks.       
An account can have many todos/tasks assigned to them, and can be set to 3 different roles:      
 - Admin  
 - User  
 - None  

An account consists of these fields:       
*AccountId*, *Name*, *Email*, *Password(Hash)*, *Salt*, *Role*, *LastAuth-Timestamp*        

A task consists of these fields:        
*TaskId*, *AccountId(Foreign-Key)*, *Created-Timestamp*, *LastUpdate-Timestamp*, *Priority*, *Task-Text*       

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
 - /auth/  
 - /tasks/  
 - /task/{taskId}  
 - /accounts/  
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

## Bugs
Totally forgot about a tasks LastUpdated field. Currently it is set freely by POST and PUT requests. Doh!       
Need to fix this.        
Same goes for the Created field I guess. It should not be allowed for the client to modify this.          

There's probably also bugs in the unit test code.        
As anyone can see by looking for example at todo_test.go, there was a lot of ugly copy&pasting involved.       
Really need to do some refactoring once I get more comfortable in Go. ;-)

## License
GPL v3, see LICENSE.TXT      



