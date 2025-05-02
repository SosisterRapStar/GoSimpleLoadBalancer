from fastapi import FastAPI
import os
from datetime import datetime
import time


app = FastAPI()

server_num = os.getenv("SERVER_NUM", "unknown")
log_path = "/logs/server.log"

@app.get("/health")
def health():
    log_entry = f"[{datetime.now()}] Health check on server #{server_num}\n"
    with open(log_path, "a") as f:
        f.write(log_entry)
    return {"status": "ok", "server": server_num}

@app.get("/long")
def longtest():
    time.sleep(30)
    return {"status": "ok", "server long/test": server_num}


@app.get("/test")
def test():
    log_entry = f"[{datetime.now()}] Test endpoint requested #{server_num}\n"
    with open(log_path, "a") as f:
        f.write(log_entry)
    return {"status": "ok", "server": server_num}



