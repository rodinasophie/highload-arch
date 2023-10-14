from random import randrange
from random import randint 
import requests
import csv 

def get_random_prefix(prefix_len):
    prefix = ""
    hexLetter = randint(0x0410, 0x042f)   
    prefix += chr(hexLetter)
    for i in range(prefix_len - 1):
        hexLetter = randint(0x0430, 0x044f)   
        prefix += chr(hexLetter)
    return prefix


def select_valid_prefix(prefix_len):
    N = 10000
    k = 0
    valid_prefix = []
    for i in range(1, N):
        firstName = get_random_prefix(prefix_len)
        secondName = get_random_prefix(prefix_len)
        url = f"http://localhost:8083/user/search?first_name={firstName}&second_name={secondName}"
        try:
            r = requests.get(url)
            if r.status_code == 200:
                valid_prefix.append([firstName, secondName])
        except requests.ConnectionError:
            print("failed to connect")
        k += 1
        if k % 100 == 0:
            print(str(k) + "/" + str(N) + " prefixes are handled, valid_prefixes[]: " + str(valid_prefix))
    return valid_prefix

valid_prefix = select_valid_prefix(3)

with open('prefix.csv', 'a', ) as f:
    print(valid_prefix)
    for item in valid_prefix:
        writer = csv.writer(f, delimiter=',')
        writer.writerow(item)