import random
import string
from locust import HttpUser, SequentialTaskSet, task
import pandas as pd
from faker import Faker
import requests

DEFAULT_PASSWORD = 'password'
PREFIX_VERSION = "/api/v2"


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
        r = send_request(method='POST', url='http://localhost:8083' + PREFIX_VERSION + '/user/register',
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
            r2 = send_request(method='POST', url='http://localhost:8083'+ PREFIX_VERSION + '/login',
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

connected_users = []

def init_dialogs(n):
    for i in range(n):
        random_users = random.sample(list(users.keys()), 2)
        user1 = random_users[0]
        user2 = random_users[1]
        connected_users.append([user1, user2, users[user2]])
        message = fake.paragraph(nb_sentences=7, variable_nb_sentences=True)
        send_request(method='POST', url = f"http://localhost:8086"+ PREFIX_VERSION + f"/dialog/" +  user1 + f"/send", json = {"text": message[:400]}, token = users[user2])
        send_request(method='POST', url = f"http://localhost:8086" + PREFIX_VERSION + f"/dialog/" +  user2 + f"/send", json = {"text": message[:400]}, token = users[user1])

init_dialogs(100)
print("Dialogs initialized.")


#for user_pair in connected_users:
#    resp = send_request('GET', url="http://localhost:8083/dialog/"+user_pair[0]+"/list", token = user_pair[2])
#    print(resp.text)

#class PerfTesting(HttpUser):
#    @task
#    def test_reads(self):
#        random_users = random.sample(list(users.keys()), 2)
#        user1 = random_users[0]
#        user2 = random_users[1]
#        self.client.get(f"/dialog/" +  user1 + f"/list", headers={"Authorization": "Bearer " + users[user2]})

