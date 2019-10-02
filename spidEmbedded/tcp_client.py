import logging
import socket
from time import sleep


def _enforce_connection(f):
    def wrapper_enforce_connection(self, *args, **kwargs):
        if not self.connected:
            m = 'Socket not connected'
            logging.critical(m)
            raise Exception(m)
        return f(self, *args, **kwargs)
    return wrapper_enforce_connection


class TCPClient:
    DEFAULT_BUFFER_SIZE = 4096
    DEFAULT_WAIT_RETRY = 3  # in seconds

    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.connected = False

        self._s = socket.socket()

    def connect(self, try_forever=False):
        self._s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        while True:
            try:
                self._s.connect((self.host, self.port))
                logging.info(f"Connected to host {self.host}:{self.port}.")
                self.connected = True
                return
            except socket.timeout as e:
                logging.critical(f"Connection to {self.host}:{self.port} timed out: `{e}`")
                if not try_forever:
                    raise e
            except (ConnectionRefusedError, OSError) as e:
                logging.critical(f"Connection refused ({self.host}:{self.port}): `{e}`")
                if not try_forever:
                    raise e
            if not try_forever:
                break
            logging.info('Retrying connection...')
            sleep(self.DEFAULT_WAIT_RETRY)

    @_enforce_connection
    def close(self):
        self._s.close()
        logging.info('Closed connection.')
        self.connected = False

    @_enforce_connection
    def send(self, message):
        logging.info(f"Sending message: `{message}`.")
        message += '\n'
        self._s.send(message.encode('ascii'))
        logging.info('Message sent.')

    @_enforce_connection
    def receive(self, size=DEFAULT_BUFFER_SIZE):
        logging.info('Waiting for message...')
        rcv = self._s.recv(size).decode('ascii')
        logging.info(f"Message received: `{rcv}`.")
        return rcv
