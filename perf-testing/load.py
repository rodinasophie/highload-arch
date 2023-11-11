import requests
count = 0
try:
    while True:
        r = requests.post('http://localhost:8083/user/register', json={"first_name": "Имя","second_name": "Фамилия",
    "birthdate": "2017-02-01","biography": "Хобби, интересы и т.п.","city": "Москва","password": "Секретная строка"})
        if r.status_code == 200:
            count += 1
        else:
            break
except:
    print(count)

print(count)


