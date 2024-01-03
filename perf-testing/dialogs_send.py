import random
import string
from locust import FastHttpUser, SequentialTaskSet, task
import pandas as pd
from faker import Faker
import requests

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
    print("Creating users\n")
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

fake = Faker('ru_RU')
Faker.seed(0)
users = {}
create_users("./people_small.csv", users)
print("Users created.")

class PerfTesting(FastHttpUser):
    @task
    def test_writes(self):
        random_users = random.sample(list(users.keys()), 2)
        user1 = random_users[0]
        user2 = random_users[1]
        message = fake.paragraph(nb_sentences=7, variable_nb_sentences=True)
        print("Send new..\n")
        resp = self.client.post(f"/dialog/" +  user1 + f"/send", json = {"text": message[:400]}, headers={"Authorization": "Bearer " + users[user2]}, )
        print("New sent..\n")