import sys
from importlib import import_module
from flask import Flask, request, jsonify
import os
import json

app = Flask(__name__)
app.debug = False
func_path = "/root/aqualake/function"
func_name = ""

@app.route('/installfunction', methods=['post'])
def install_func():
    resp = {
        "Ok": True,
        "Err": ""
    }

    if not request.data:
        resp["Ok"] = False
        resp["Err"] = "Request Body Err"
        response = app.response_class(
            response=json.dumps(resp),
            status=200,
            mimetype='application/json'
        )
        return response

    content = request.get_json()

    global func_name
    func_name = content['Name']
    url = content['Url']
    
    os.popen('wget --auth-no-challenge --user admin --password 1234  -O /root/aqualake/function.py ' + url).read()
    os.popen('pipreqs /root/aqualake')
    os.popen('pip install -r /root/aqualake/requirements.txt')

    response = app.response_class(
        response=json.dumps(resp),
        status=200,
        mimetype='application/json'
    )
    return response

@app.route('/trigger', methods=['post'])
def trigger():
    resp = {
        "Ret": "",
        "Err": ""
    }

    content = request.get_json()
    if "Args" not in content:
        resp["Err"] = "No Args in Body"
        response = app.response_class(
            response=json.dumps(resp),
            status=200,
            mimetype='application/json'
        )
        return response

    args = content["Args"]

    module = import_module("function")
    function = getattr(module, func_name)

    resp["Ret"] = function(*args)
    response = app.response_class(
        response=json.dumps(resp),
        status=200,
        mimetype='application/json'
    )
    return response


if __name__ == '__main__':
    ip = os.popen("cat /etc/hosts | awk 'END{print $1}'").read()
    ip = ip.replace('\n','\0')
    # print("**", ip, "**")
    app.run(host=ip, port=8698)
