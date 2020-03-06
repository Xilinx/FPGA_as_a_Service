import socket
import sys
import os
import subprocess
import datetime
from threading import Thread

SERVER_IP = '10.96.59.3'
SERVER_PORT = 8010
FPGA_CMD = '/opt/xilinx/k8s/server/fpga_host_exe'
FPGA_ARG = '/opt/xilinx/k8s/server/fpga_algo.awsxclbin'

class MyServer():

    def __init__(self):
        self.server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.server.bind(('0.0.0.0', SERVER_PORT))
        self.server.listen(5)

    def handle_connection(self, client_socket):
        req = client_socket.recv(4096)
        # run fpga helloworld, send the output back
	out = subprocess.Popen([FPGA_CMD, FPGA_ARG], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
	stdout,stderr = out.communicate()
        client_socket.send("Response from FPGA server:\n" + stdout)
        client_socket.close()

    def start(self):
        print "Waiting for command from the client..."
        while True:
            client_sock, address = self.server.accept()
            client_handler = Thread(target=self.handle_connection, args=(client_sock,))
            client_handler.start()


def client_send():
    bind_ip = SERVER_IP
    bind_port = SERVER_PORT
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client.connect((bind_ip, bind_port))
    client.send("hello")
    response = client.recv(4096)
    client.close()
    print "Send request to server..."	
    print response
    print "--END--"

def usage():
    print("Usage:")
    print("\t%s " % sys.argv[0])

def main():
    if (len(sys.argv) != 1):
        usage()
        exit(0)
    print "Started FPGA Server Version 1.0: ",datetime.datetime.now()
    server = MyServer()
    server.start()
main()

