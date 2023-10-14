import random
import string
from locust import HttpUser, task
import pandas as pd
from locust_plugins.csvreader import CSVReader

prefix_reader = CSVReader("prefix.csv")

class PerfTesting(HttpUser):
    @task
    def perf_testing(self):
        prefix = next(prefix_reader)
        firstName = prefix[0]
        secondName = prefix[1]
        self.client.get(f"/user/search?first_name={firstName}&second_name={secondName}")

