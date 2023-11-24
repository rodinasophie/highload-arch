import random
from faker import Faker
import csv

DEFAULT_UUID = '03881183-2362-41c1-b4cd-7552724cdb33'

fake = Faker('ru_RU')
Faker.seed(0)
with open("./db-data-leader/posts.csv", "w") as csv_file:
    for _ in range(100):
        text = fake.paragraph(nb_sentences=7, variable_nb_sentences=True)
        created_at = fake.date_time_this_year()
        updated_at = fake.date_time_between(start_date=created_at)
        line = '\t'.join([DEFAULT_UUID, text, created_at.strftime('%Y-%m-%d %H:%M:%S'), updated_at.strftime('%Y-%m-%d %H:%M:%S')])
        csv_file.write(line)
        csv_file.write('\n')