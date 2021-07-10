import asyncio
import builtins
import time

import orjson

from modules.core import *
from lynxfall.rabbit.core import *


class PIDRecorder():
    def __init__(self):
        self.pids = []
        self.session_id = None

    def record(self, pid):
        if pid in self.pids:
            return
        self.pids.append(pid)
        self.pids.sort()

    def remove(self, pid):
        try:
            pid_index = self.pids.index(pid)
        except ValueError:
            return
        del self.pids[pid_index]

    def worker_amt(self):
        return len(self.pids)

    def reset(self):
        self.pids = []

    def list(self):
        return self.pids

async def status(state, pidrec):
    pubsub = state.redis.pubsub()
    await pubsub.subscribe(f"_worker")
    flag = True
    async for msg in pubsub.listen():
        if flag:
            await state.redis.publish(f"_worker", "NOSESSION UP RMQ 0") # Announce that we are up
            flag = False
        print(msg)
        if msg is None or type(msg.get("data")) != bytes:
            continue
        msg = tuple(msg.get("data").decode("utf-8").split(" "))
        
        match msg:
            case (session_id, "UP", ("RMQ" | "WORKER") as tgt, pid, reload, worker_amt) if pid.isdigit() and reload.isdigit() and worker_amt.isdigit():
            
                logger.info(f"{tgt} {pid} is now up with reload mode {reload}. Amount of workers is {worker_amt}")

                if not pidrec.session_id:
                    pidrec.session_id = session_id

                elif session_id != pidrec.session_id:
                    # Assume new session
                    logger.info("Made new worker session due to new session state")
                    pidrec.reset()
                    pidrec.session_id = session_id

                pidrec.record(int(pid)) if tgt == "WORKER" else None

                if pidrec.worker_amt() > int(worker_amt) and tgt == "WORKER":
                    logger.warning(f"Invalid worker {worker_amt} with pid {pid} added. Restting config and publishing REGET")
                    pidrec.reset()
                    await asyncio.sleep(1)
                    await state.redis.publish(f"_worker", f"{pidrec.session_id} REGET WORKER INVALID_STATE")

                if pidrec.worker_amt() == int(worker_amt) and tgt == "WORKER":
                    logger.success("All workers are now up")
                    await asyncio.sleep(1)
                    worker_pids = " ".join([str(pid) for pid in pidrec.list()])
                    await state.redis.publish(f"_worker", f"{pidrec.session_id} FUP {worker_pids}")

            case ("DOWN", ("RMQ" | "WORKER") as tgt, pid) if pid.isdigit():
                logger.info(f"{tgt} {pid} is now down")
                pidrec.remove(int(pid)) if tgt == "RMQ" else None
            
            case _:
                logger.warning(f"Unhandled message {msg}")

                
async def prehook(config, *, state):
    builtins.pidrec = PIDRecorder()
    asyncio.create_task(status(state, pidrec))


async def backend(state, json, *args, **kwargs):
    return pidrec.list()


class Config:   
    queue = "_worker"
    name = "Worker Handler" 
    descriprion = "Handle Workers (PIDs right now but may be increased in future)"
    pre = prehook
