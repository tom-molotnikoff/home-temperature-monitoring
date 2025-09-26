from flask import Flask, jsonify
import random
import datetime

app = Flask(__name__)

@app.route('/temperature')
def temperature():
    temp = round(random.uniform(18.0, 22.0), 2)
    now = datetime.datetime.now(datetime.timezone.utc)
    formatted = now.strftime("%Y-%m-%d %H:%M:%S")
    return jsonify({"temperature": temp, "time": formatted})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)