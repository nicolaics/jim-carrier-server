import requests as rq
import random

def create_users(BACKEND_ROOT):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Adam adam",
            "email": "adam1@gmail.com",
            "password": "adampassword",
            "phoneNumber": "0819288176",
            "verificationCode": "123456",
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Bob bob",
            "email": "bob@gmail.com",
            "password": "bobpassword",
            "phoneNumber": "0819288326",
            "verificationCode": "123456",
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Charlie charlie",
            "email": "charlie@gmail.com",
            "password": "charliepassword",
            "phoneNumber": "0813318326",
            "verificationCode": "123456",
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Delta delta",
            "email": "delta@gmail.com",
            "password": "deltapassword",
            "phoneNumber": "081921296",
            "verificationCode": "123456",
        },
    )
    print(r.json())

def login_users(BACKEND_ROOT):
    tokens = []

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/login",
        json={
            "email": "adam1@gmail.com",
            "password": "adampassword",
        },
    )
    print(r.json())
    tokens.append(r.json()['token'])

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/login",
        json={
            "email": "bob@gmail.com",
            "password": "bobpassword",
        },
    )
    print(r.json())
    tokens.append(r.json()['token'])

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/login",
        json={
            "email": "charlie@gmail.com",
            "password": "charliepassword",
        },
    )
    print(r.json())
    tokens.append(r.json()['token'])

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/login",
        json={
            "email": "delta@gmail.com",
            "password": "deltapassword",
        },
    )
    print(r.json())
    tokens.append(r.json()['token'])
    
    return tokens

def create_listings(BACKEND_ROOT, tokens):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/listing",
        json={
            "destination": "Alabama",
            "weightAvailable": 20,
            "pricePerKg": 12000,
            "currency": "KRW",
            "departureDate": "2024-10-09 +0900KST",
            "lastReceivedDate": "2024-10-05 +0900KST",
            "bankName": "abc",
            "accountNumber": "1234",
            "accountHolder": "pweo",
        },
        headers={
            "Authorization": "Bearer " + tokens[0],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/listing",
        json={
            "destination": "Jakarta",
            "weightAvailable": 50.0,
            "pricePerKg": 11500.0,
            "currency": "IDR",
            "departureDate": "2024-12-09 +0700KST",
            "lastReceivedDate": "2024-11-05 +0700UTC",
            "bankName": "abc",
            "accountNumber": "1234",
            "accountHolder": "pweo",
        },
        headers={
            "Authorization": "Bearer " + tokens[1],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/listing",
        json={
            "destination": "Malaysia",
            "weightAvailable": 22,
            "pricePerKg": 11000,
            "currency": "MYR",
            "departureDate": "2025-10-09 +0900KST",
            "lastReceivedDate": "2025-10-05 +0900KST",
            "bankName": "abc",
            "accountNumber": "1234",
            "accountHolder": "pweo",
        },
        headers={
            "Authorization": "Bearer " + tokens[2],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/listing",
        json={
            "destination": "Nepal",
            "weightAvailable": 15,
            "pricePerKg": 10000,
            "currency": "USD",
            "departureDate": "2025-01-10 +0900KST",
            "lastReceivedDate": "2025-01-05 +0900KST",
            "bankName": "abc",
            "accountNumber": "1234",
            "accountHolder": "pweo",
        },
        headers={
            "Authorization": "Bearer " + tokens[3],
        },
    )
    print(r.json())

def create_orders(BACKEND_ROOT, tokens):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/order",
        json={
            "listingId": 1,
            "weight": 5,
            "price": (12000 * 5),
            "currency": "KRW",
            "packageContent": "Test",
            "packageImage": "/9j/4AAQSkZJRgABAQABAAD/2wBD"
        },
        headers={
            "Authorization": "Bearer " + tokens[0],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/order",
        json={
            "listingId": 2,
            "weight": 2,
            "price": (11000 * 2),
            "currency": "IDR",
            "packageContent": "Test",
            "packageImage": "/9j/4AAQSkZJRgABAQABAAD/2wBD",
            "noted": "test",
        },
        headers={
            "Authorization": "Bearer " + tokens[1],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/order",
        json={
            "listingId": 3,
            "weight": 8,
            "price": (11000 * 3),
            "packageContent": "Test",
            "packageImage": "/9j/4AAQSkZJRgABAQABAAD/2wBD",
            "currency": "KRW",
        },
        headers={
            "Authorization": "Bearer " + tokens[2],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/order",
        json={
            "listingId": 4,
            "weight": 5,
            "price": (12000 * 5),
            "currency": "USD",
            "packageContent": "Test",
            "packageImage": "/9j/4AAQSkZJRgABAQABAAD/2wBD",
        },
        headers={
            "Authorization": "Bearer " + tokens[3],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/order",
        json={
            "listingId": 4,
            "weight": 26,
            "price": (12000 * 26),
            "currency": "USD",
            "packageContent": "Test",
            "packageImage": "/9j/4AAQSkZJRgABAQABAAD/2wBD",
        },
        headers={
            "Authorization": "Bearer " + tokens[2],
        },
    )
    print(r.json())

def create_reviews(BACKEND_ROOT, tokens):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/review",
        json={
            "orderId": 1,
            "revieweeName": "Bob bob",
            "content": "nice",
            "rating": 5,
        },
        headers={
            "Authorization": "Bearer " + tokens[0],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/review",
        json={
            "orderId": 2,
            "revieweeName": "Adam adam",
            "rating": 4,
        },
        headers={
            "Authorization": "Bearer " + tokens[1],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/review",
        json={
            "orderId": 3,
            "revieweeName": "Adam adam",
            "content": "kind",
            "rating": 4,
        },
        headers={
            "Authorization": "Bearer " + tokens[2],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/review",
        json={
            "orderId": 4,
            "revieweeName": "Adam adam",
            "content": "okay",
            "rating": 3,
        },
        headers={
            "Authorization": "Bearer " + tokens[3],
        },
    )
    print(r.json())

def logout_users(BACKEND_ROOT, tokens):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/logout",
        headers={
            "Authorization": "Bearer " + tokens[0],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/logout",
        headers={
            "Authorization": "Bearer " + tokens[1],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/logout",
        headers={
            "Authorization": "Bearer " + tokens[2],
        },
    )
    print(r.json())

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/logout",
        headers={
            "Authorization": "Bearer " + tokens[3],
        },
    )
    print(r.json())

def main():
    # BACKEND_ROOT = input("enter backend host:port: ")
    BACKEND_ROOT = "localhost:9988"
    create_users(BACKEND_ROOT)
    tokens = login_users(BACKEND_ROOT)
    create_listings(BACKEND_ROOT, tokens)
    create_orders(BACKEND_ROOT, tokens)
    create_reviews(BACKEND_ROOT, tokens)
    logout_users(BACKEND_ROOT, tokens)
    

    print("DONE")


if __name__ == "__main__":
    main()
