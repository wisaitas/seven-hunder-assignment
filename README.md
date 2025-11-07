# prepare

1. start docker local

2. run

```
docker-compose up -d
```

3. install extension `REST Client`

4. all api in folder httptest

# test register

- click register.http
- call register 2 users

# test login

- click login.http
- use email and password api register for login click login.http

# test get users

- click getusers.http
- call login api
- call getusers api

# test get user by id

- copy id from api get users
- click getuserbyid.http
- replace parameter path
- call login api
- call getuserbyid api

# test update user

- copy id from api get users
- click updateuser.http
- replace parameter path
- call login api
- call updateuser api

# test delete user (soft delete)

- copy id from api get users
- click deleteuser.http
- replace parameter path
- call login api
- call deleteuser api

# test create user

- click createuser.http
- call login api
- call createuser api
