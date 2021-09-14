from fastapi import Depends

from .. import db


async def reset_db_state():
    db.db._state._state.set(db.db_state_default.copy())
    db.db._state.reset()


def get_db(db_state=Depends(reset_db_state)):
    try:
        db.db.connect()
        yield
    finally:
        if not db.db.is_closed():
            db.db.close()
