import requests
import pandas as pd
from faker import Faker
import random
import time

users = {}

DEFAULT_PASSWORD = 'password'

def send_request(method, url, token="", json={}):
    if token != "":
        headers = {"Authorization": "Bearer " + token}
    else:
        headers = {}
    if method == 'GET':
        return requests.get(url, headers = headers)
    elif method == 'POST':
        return requests.post(url, headers = headers, json = json)
    elif method == 'PUT':
        return requests.put(url, headers = headers, json = json)

def create_users(filename, users):
    df = pd.read_csv(filename, delimiter='\t', header=None, keep_default_na=False)
    for row in df.itertuples():
        r = send_request(method='POST', url='http://localhost:8083/user/register',
                    json = {"first_name": row[1],
                            "second_name": row[2],
                            "birthdate": row[3],
                            "biography": row[5],
                            "city": row[4],
                            "password": DEFAULT_PASSWORD
                            })
        if r.status_code != 200:
            print(r)
        else:
            r2 = send_request(method='POST', url='http://localhost:8083/login',
                           json = {
                              "id": r.json()['user_id'],
                              "password": DEFAULT_PASSWORD
                           })
            if r2.status_code != 200:
                print(r2)
            else:
                users[r.json()['user_id']] = r2.json()['token']

def make_friends(users, minFriends=0, maxFriends=20):
    # Each friend creates from 0 to 20 random connections
    ids = users.keys()
    for id in ids:
        numberOfFriends = random.randint(minFriends, maxFriends)
        friends = random.sample(list(ids), numberOfFriends)
        for friend in friends:
            if friend != id:
                r = send_request(method='PUT', url='http://localhost:8083/friend/add/' + friend, token=users[id])
                if r.status_code != 200:
                    print(r)


def create_posts(users, fake, minPosts=0, maxPosts=20):
    ids = users.keys()
    numberOfPosts = random.randint(minPosts, maxPosts)
    for id in ids:
        for n in range(numberOfPosts):
            post = fake.paragraph(nb_sentences=7, variable_nb_sentences=True)
            create_post(id, users[id], post)

def create_post(userID, token, post):
    r = send_request(method='POST', url='http://localhost:8083/post/create', json = {"text": post}, token=token)
    if r.status_code != 200:
        print(r)

def feed_posts(users):
    numberOfUsers = len(users)
    randomUsers = random.sample(list(users.keys()), 1)
    r = send_request(method='GET', url='http://localhost:8083/post/feed?offset=0&limit=2', token=users[randomUsers[0]])
    if r.status_code != 200:
        print(r)
    else:
        print(r.json())


def main():
    fake = Faker('ru_RU')
    Faker.seed(0)

    create_users("./people_small.csv", users)
    make_friends(users, 0, 20)
    create_posts(users, fake, 0, 10)
    time.sleep(60)
    feed_posts(users)

main()