import time

import requests
import urllib3
from datetime import datetime

urllib3.disable_warnings()

base_url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pg/{}'
headers = {
    'Host': 'us-east-2.pg-client.konghq.tech',
}


class JSONResponse:
    def __init__(self, status_code, headers, payload):
        self.statusCode = status_code
        self.headers = headers
        self.payload = payload

    def __str__(self):
        return '{} - code={}, payload={}'.format(datetime.now(), self.statusCode, self.payload)


def get_main_replication_status():
    res = requests.request("GET", base_url.format('pghealth'), headers=headers, verify=False)
    return JSONResponse(res.status_code, res.headers, res.json())


def get_ro_replication_status():
    res = requests.request("GET", base_url.format('replstatusro'), headers=headers, verify=False)
    return JSONResponse(res.status_code, res.headers, res.json())


def get_foo():
    res = requests.request("GET", base_url.format('foo'), headers=headers, verify=False)
    return JSONResponse(res.status_code, res.headers, res.json())


def poolstats():
    res = requests.request("GET", base_url.format('poolstats'), headers=headers, verify=False)
    return JSONResponse(res.status_code, res.headers, res.json())


def post_foo():
    res = requests.request("POST", base_url.format('foo'), headers=headers, verify=False)
    return JSONResponse(res.status_code, res.headers, res.json())


def make_pg_calls():
    response = get_main_replication_status()
    print('main     {}'.format(response))
    response = get_ro_replication_status()
    print('ro       {}'.format(response))
    response = poolstats()
    print('pool     {}'.format(response))
    response = get_foo()
    print('getfoo   {}'.format(response))
    response = post_foo()
    print('postfoo  {}'.format(response))


if __name__ == '__main__':
    print('starting pgx run..')
    for i in range(1, 3):
        make_pg_calls()
        print('==')
        time.sleep(2)
