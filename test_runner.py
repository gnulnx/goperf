#!/usr/bin/env python
import json
import requests

try:
    from json.decoder import JSONDecodeError
except Exception:
    from simplejson import JSONDecodeError

for conn in range(2, 200, 2):

    r = requests.post("http://127.0.0.1:9000/api/", {
        'url': 'https://stage2.teaquinox.com/',
        'seconds': 10,
        'conn': conn
    })

    if r.status_code == 201:
        resp_time = r.json()['base_url']['avg_page_resp_time'] * 0.000000001
        first_byte = r.json()['base_url']['avg_time_to_first_byte'] * 0.000000001
        num_requests = r.json()['base_url']['num_reqs']
        status = r.json()['base_url']['status']

        #print(json.dumps(r.json(), indent=4, sort_keys=True, default=str))
        msg = "%s, %s, %s, %s, %s" % (
            conn, "%.3f" % resp_time, "%.3f" % first_byte, num_requests, status
        )
        print(msg)
    else:
        print(r.status_code)
