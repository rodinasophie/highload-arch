import requests
import pandas as pd
from faker import Faker
import random
import time
from websocket import create_connection
API_PREFIX = '/api/v2'

BACKEND_LOCAL_URL = 'http://localhost:8083'
DIALOGS_LOCAL_URL = 'http://localhost:8087'
COUNTERS_LOCAL_URL = 'http://localhost:8091'

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
        r = send_request(method='POST', url=BACKEND_LOCAL_URL+API_PREFIX+'/user/register',
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
            r2 = send_request(method='POST', url=BACKEND_LOCAL_URL+API_PREFIX+'/login',
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
                r = send_request(method='PUT', url=BACKEND_LOCAL_URL+API_PREFIX+'/friend/add/' + friend, token=users[id])
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
    r = send_request(method='POST', url=BACKEND_LOCAL_URL+API_PREFIX+'/post/create', json = {"text": post}, token=token)
    if r.status_code != 200:
        print(r)

def feed_posts(users):
    numberOfUsers = len(users)
    randomUsers = random.sample(list(users.keys()), 1)
    r = send_request(method='GET', url=BACKEND_LOCAL_URL+API_PREFIX+'/post/feed?offset=0&limit=2', token=users[randomUsers[0]])
    if r.status_code != 200:
        print(r)
    else:
        print(r.json())

def create_dialogs(users, fake, minMessages=0, maxMessages=20, numberOfDialogPairs=10):
    dialogs = []
    numberOfMessages = random.randint(minMessages, maxMessages)
    for i in range(numberOfDialogPairs):
        user_pairs = random.sample(list(users.keys()), 2)
        user1 = user_pairs[0]
        user2 = user_pairs[1]
        dialogs.append(user_pairs)
        for k in range(numberOfMessages):
            message = fake.paragraph(nb_sentences=7, variable_nb_sentences=True)
            response = fake.paragraph(nb_sentences=7, variable_nb_sentences=True)
            send_message(user1, users[user2], message)
            send_message(user1, users[user2], message)
            send_message(user2, users[user1], response)
    return dialogs

def send_message(recepientID, token, message):
    r = send_request(method='POST', url=DIALOGS_LOCAL_URL+API_PREFIX+'/dialog/' + recepientID + '/send', json = {"text": message}, token=token)
    if r.status_code != 200:
        print(r, recepientID, token, message)

def list_dialogs(users, dialogs):
    for d in dialogs:
        user1 = d[0]
        user2 = d[1]
        send_list_dialogs(user1, users[user2])
        send_list_dialogs(user2, users[user1])

def send_list_dialogs(recepientID, token):
    r = send_request(method='GET', url=DIALOGS_LOCAL_URL+API_PREFIX+'/dialog/' + recepientID + '/list', token=token)
    if r.status_code != 200:
        print(r, recepientID, token)

def main():
    fake = Faker('ru_RU')
    Faker.seed(0)

    create_users("./people_small.csv", users)
    make_friends(users, 0, 20)
    #create_posts(users, fake, 0, 10)
    #time.sleep(60)
    #feed_posts(users)
    dialogs = create_dialogs(users, fake, 0, 3, 5)
    list_dialogs(users, dialogs)
main()