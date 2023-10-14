import random
from faker import Faker
import csv

fake = Faker('ru_RU')
Faker.seed(0)
with open("people.csv", "w") as csv_file:
    for _ in range(1000000):
        if random.randint(0, 1):
            first_name = fake.first_name_female()
            last_name = fake.last_name_female()
        else:
            first_name = fake.first_name_male()
            last_name = fake.last_name_male()
        line = '\t'.join([first_name, last_name, fake.date_of_birth().strftime('%Y-%m-%d'), fake.city(), ""])
        csv_file.write(line)
        csv_file.write('\n')