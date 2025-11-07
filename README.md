# step run

1. start docker

2. docker-compose up -d

3, install extension `REST Client`

4. all api in folder httptest

# test create user (register)

click register.http click create user atleast 2 users

# test get users

click getusers.http call api

# test get user by id

use data api get users copy id for call this api click getuserbyid.http replace user_id parameter then call api

# test update user

- use data api get users copy id replace user_id parameter updateuser.http then call api
- test get users again see result change after update

# test delete user (soft delete)

- use data api get users copy id replace user_id parameter deleteuser.http then call api
- test get users again see result change after delete
